package sigmon

import (
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	sigs = []syscall.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	}
)

type checkable struct {
	sync.Mutex
	id  int
	val int
	ct  int
}

func (c *checkable) handler(sm *SignalMonitor) {
	c.Lock()
	defer c.Unlock()

	c.val = c.id
	c.ct++
}

func (c *checkable) info() (id, val, ct int) {
	c.Lock()
	defer c.Unlock()

	return c.id, c.val, c.ct
}

func TestUnitSignalJunctionConnect(t *testing.T) {
	j := newSignalJunction()
	j.connect()

	for _, s := range sigs {
		if err := callOSSignal(s); err != nil {
			t.Errorf("unexpected error when calling %s: %s", s, err)
		}
	}

	if !receiveOnAll(j) {
		t.Fatal("should not wait forever")
	}
}

func TestUnitSignalJunctionDisconnect(t *testing.T) {
	j := newSignalJunction()
	j.connect()
	j.disconnect()

	if receiveOnAll(j) {
		t.Fatal("should wait forever")
	}
}

func TestUnitSignalHandlerRegister(t *testing.T) {
	c1 := &checkable{id: 123}
	c2 := &checkable{id: 234}

	h := newSignalHandler(nil)
	h.register(c1.handler)
	h.register(c2.handler)

	select {
	case fn := <-h.registry:
		if fn == nil {
			t.Error("want function, got nil")
		}

		fn(&SignalMonitor{})
	case <-time.After(time.Millisecond):
		t.Error("should not wait forever")
	}

	_, c1Val, _ := c1.info()
	if 0 != c1Val {
		t.Errorf("want %d, got %d", 0, c1Val)
	}
	c2ID, c2Val, _ := c2.info()
	if c2ID != c2Val {
		t.Errorf("want %d, got %d", c2ID, c2Val)
	}
}

func TestUnitSignalHandlerSet(t *testing.T) {
	c := &checkable{id: 123}
	h := newSignalHandler(nil)
	h.set(c.handler)

	h.handler(&SignalMonitor{})

	id, val, _ := c.info()
	if id != val {
		t.Errorf("want %d, got %d", id, val)
	}
}

func TestUnitSignalHandlerHandle(t *testing.T) {
	c := &checkable{id: 123}
	h := newSignalHandler(c.handler)

	h.handle(&SignalMonitor{})

	id, val, _ := c.info()
	if id != val {
		t.Errorf("want %d, got %d", id, val)
	}
}

func TestUnitSignalMonitorSet(t *testing.T) {
	c := &checkable{id: 123}
	m := New(nil)
	m.Set(c.handler)

	select {
	case fn := <-m.handler.registry:
		if fn == nil {
			t.Error("want function, got nil")
		}
	case <-time.After(time.Millisecond):
		t.Error("should not wait forever")
	}
}

func TestUnitSignalMonitorRun(t *testing.T) {
	c := &checkable{id: 123}
	m := New(c.handler)
	if m.on {
		t.Errorf("want %t, got %t", false, m.on)
	}

	m.Run()
	m.Run()
	if !m.on {
		t.Errorf("want %t, got %t", true, m.on)
	}

	s := syscall.SIGHUP
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	id, val, ct := c.info()
	if id != val {
		t.Errorf("want %d, got %d", id, val)
	}
	if ct > 1 {
		t.Error("signal possibly connected multiple times")
	}
}

func receiveOnAll(j *signalJunction) bool {
	for i := 0; i < 5; i++ {
		select {
		case <-j.sighup:
		case <-j.sigint:
		case <-j.sigterm:
		case <-j.sigusr1:
		case <-j.sigusr2:
		case <-time.After(time.Millisecond):
			return false
		}
	}

	return true
}

func callOSSignal(s syscall.Signal) error {
	if err := syscall.Kill(syscall.Getpid(), s); err != nil {
		return err
	}

	// delay for requested signal propagation
	for i := 1 << 13; i > 0; i-- {
		syscall.Getpid()
	}

	return nil
}
