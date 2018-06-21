// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

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

func (c *checkable) handler(*State) {
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
	j.disconnect()

	if receiveOnAll(j) {
		t.Fatal("should wait forever")
	}
}

func TestUnitSignalHandlerRegister(t *testing.T) {
	c1 := &checkable{id: 123}
	c2 := &checkable{id: 234}

	h := newHandlerFuncRegistry(nil)
	h.load(c1.handler)
	h.load(c2.handler)

	select {
	case fn := <-h.reg:
		if fn == nil {
			t.Error("got nil, want not nil")
		}

		fn(&State{})
	case <-time.After(time.Millisecond):
		t.Error("should not wait forever")
	}

	_, c1Val, _ := c1.info()
	if c1Val != 0 {
		t.Errorf("got %d, want %d", c1Val, 0)
	}
	c2ID, c2Val, _ := c2.info()
	if c2Val != c2ID {
		t.Errorf("got %d, want %d", c2Val, c2ID)
	}
}

func TestUnitSignalHandlerSet(t *testing.T) {
	c := &checkable{id: 123}
	h := newHandlerFuncRegistry(nil)
	h.crank(c.handler)

	h.fn(&State{})

	id, val, _ := c.info()
	if val != id {
		t.Errorf("got %d, want %d", val, id)
	}
}

func TestUnitSignalHandlerHandle(t *testing.T) {
	c := &checkable{id: 123}
	h := newHandlerFuncRegistry(c.handler)

	h.handle(&State{})

	id, val, _ := c.info()
	if val != id {
		t.Errorf("got %d, want %d", val, id)
	}
}

func TestUnitSignalMonitorSet(t *testing.T) {
	c := &checkable{id: 123}
	m := New(nil)
	m.Set(c.handler)

	select {
	case fn := <-m.r.reg:
		if fn == nil {
			t.Error("got nil, want not nil")
		}
	case <-time.After(time.Millisecond):
		t.Error("should not wait forever")
	}
}

func TestUnitSignalMonitorPreScan(t *testing.T) {
	m := New(nil)

	m.r.reg = make(chan HandlerFunc, 1)
	m.r.reg <- func(*State) {}

	got := m.preScan()
	want := true
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}

	select {
	case <-m.r.reg:
		t.Error("failed to read from channel")
	default:
	}

	m.done = make(chan struct{}, 1)
	m.done <- struct{}{}

	got = m.preScan()
	want = false
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}

	got = m.preScan()
	want = true
	if got != want {
		t.Errorf("got %t, want %t", got, want)
	}
}

func TestUnitSignalMonitorScan(t *testing.T) {
	m := New(nil)

	time.AfterFunc(time.Second*6, func() {
		go func() {
			m.done <- struct{}{}
		}()

		got, want := m.scan(), false
		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}

		go func() {
			m.r.reg <- func(*State) {}
			m.j.sighup <- syscall.SIGHUP
			m.j.sigint <- syscall.SIGINT
			m.j.sigterm <- syscall.SIGTERM
			m.j.sigusr1 <- syscall.SIGUSR1
			m.j.sigusr2 <- syscall.SIGUSR2
		}()

		for i := 0; i < 6; i++ {
			got, want := m.scan(), true
			if got != want {
				t.Errorf("got %t, want %t", got, want)
			}
		}
	})
}

func TestUnitSignalMonitorMonitor(t *testing.T) {
	m := New(nil)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go m.monitor(wg)

	m.r.reg <- func(s *State) {
		m.done <- struct{}{}
	}
	m.j.sighup <- syscall.SIGHUP

	select {
	case <-m.j.sighup:
	default:
		t.Error("signal should not have been handled")
	}
}

func TestUnitSignalMonitorStart(t *testing.T) {
	c := &checkable{id: 123}
	m := New(c.handler)
	if m.isOn {
		t.Errorf("got %t, want %t", m.isOn, false)
	}

	m.Start()
	m.Start()
	if !m.isOn {
		t.Errorf("got %t, want %t", m.isOn, true)
	}

	s := syscall.SIGHUP
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	m.Stop()

	id, val, ct := c.info()
	if val != id {
		t.Errorf("got %d, want %d", val, id)
	}
	if ct > 1 {
		t.Error("signal possibly connected multiple times")
	}
}

func TestUnitSignalMonitorStop(t *testing.T) {
	c := &checkable{id: 123}
	m := New(c.handler)
	m.Start()

	s := syscall.SIGHUP
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	m.Stop()

	if m.isOn {
		t.Errorf("got %t, want %t", m.isOn, false)
	}

	mx := New(nil)
	mx.Start()

	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	mx.Stop()

	_, _, ct := c.info()
	if ct != 1 {
		t.Errorf("got %d, want %d", ct, 1)
	}
}

func TestUnitSignalMonitorSig(t *testing.T) {
	m := New(nil)
	m.setState(SIGHUP)

	got, want := m.s.Signal(), SIGHUP
	if got != want {
		t.Errorf("got %s, want %s", got, want)
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
		case <-time.After(time.Microsecond * 100):
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
	for i := 1 << 23; i > 0; i-- {
	}

	return nil
}
