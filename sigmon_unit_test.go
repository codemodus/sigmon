package sigmon

import (
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
	id  int
	val int
}

func (c *checkable) handler(sm *SignalMonitor) {
	c.val = c.id
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

	if 0 != c1.val {
		t.Errorf("want %d, got %d", 0, c1.val)
	}
	if c2.id != c2.val {
		t.Errorf("want %d, got %d", c2.id, c2.val)
	}
}

func TestUnitSignalHandlerSet(t *testing.T) {
	c := &checkable{id: 123}
	h := newSignalHandler(nil)
	h.set(c.handler)

	h.handler(&SignalMonitor{})

	if c.id != c.val {
		t.Errorf("want %d, got %d", c.id, c.val)
	}
}

func TestUnitSignalHandlerHandle(t *testing.T) {
	c := &checkable{id: 123}
	h := newSignalHandler(c.handler)

	h.handle(&SignalMonitor{})

	if c.id != c.val {
		t.Errorf("want %d, got %d", c.id, c.val)
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
	return syscall.Kill(syscall.Getpid(), s)
}
