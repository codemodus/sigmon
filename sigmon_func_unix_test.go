// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon_test

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/codemodus/sigmon"
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

func (c *checkable) handler(sm *sigmon.SignalMonitor) {
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

func TestFuncSignalIgnorance(t *testing.T) {
	sm := sigmon.New(nil)
	sm.Start()

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}
}

func TestFuncSignalConstantRetrieval(t *testing.T) {
	sm := sigmon.New(nil)
	sm.Start()

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	got, want := sm.Sig(), sigmon.SIGINT
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFuncSignalMonitorDoubleSetAndStop(t *testing.T) {
	c := &checkable{id: 123}
	sm := sigmon.New(nil)

	sm.Set(c.handler)
	sm.Set(c.handler)

	sm.Start()
	sm.Stop()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	_, _, got := c.info()
	want := 0
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestFuncSignalHandling(t *testing.T) {
	c := &checkable{id: 123}

	sm := sigmon.New(nil)
	sm.Start()
	sm.Set(c.handler)

	for i, s := range sigs {
		if err := callOSSignal(s); err != nil {
			t.Errorf("unexpected error when calling %s: %s", s, err)
		}

		time.AfterFunc(time.Second*6, func() {
			_, _, got := c.info()
			want := i
			if got != want {
				t.Errorf("got %d, want %d", got, want)
			}
		})
	}
}

func callOSSignal(s syscall.Signal) error {
	if err := syscall.Kill(syscall.Getpid(), s); err != nil {
		return err
	}

	// delay for requested signal propagation
	delay()

	return nil
}

func delay() {
	for i := 1 << 23; i > 0; i-- {
	}
}
