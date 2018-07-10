// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon

import (
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"testing"
)

func TestSignalMonitor(t *testing.T) {
	t.Run("SignalSuppression", tSignalMonitorSignalSuppression)
	t.Run("ConstantRetrieval", tSignalMonitorConstantRetrieval)
	t.Run("ExtraMethodCalls", tSignalMonitorExtraMethodCalls)
	t.Run("SignalHandling", tSignalMonitorSignalHandling)
}

func tSignalMonitorSignalSuppression(t *testing.T) {
	sm := New(nil)
	sm.Start()

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	sm.Stop()
}

func tSignalMonitorConstantRetrieval(t *testing.T) {
	ct := int32(0)
	f := func(s *State) {
		atomic.AddInt32(&ct, 1)

		got, want := s.Signal(), SIGINT
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}

	sm := New(f)
	sm.Start()

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	sm.Stop()

	runtime.Gosched()

	if atomic.LoadInt32(&ct) == 0 {
		t.Errorf("handlerfunc not called")
	}
}

func tSignalMonitorExtraMethodCalls(t *testing.T) {
	ct := 0
	f := func(*State) { ct++ }
	sm := New(nil)
	sm.Stop()
	sm.Stop()
	sm.Start()
	sm.Stop()

	sm.Set(f)
	sm.Set(f)
	sm.Start()
	sm.Start()
	sm.Set(f)
	sm.Set(f)
	sm.Stop()
	sm.Stop()

	runtime.Gosched() // ensure sm loop spins

	c := make(chan os.Signal, 1)
	defer close(c)
	signal.Notify(c, syscall.SIGINT)
	defer signal.Stop(c)

	s := syscall.SIGINT
	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	got := ct
	want := 0
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func tSignalMonitorSignalHandling(t *testing.T) {
	ct := int32(-1)
	f := func(*State) { atomic.AddInt32(&ct, 1) }
	sm := New(f)
	sm.Start()

	for i, s := range osSignals {
		if err := callOSSignal(s); err != nil {
			t.Errorf("unexpected error when calling %s: %s", s, err)
		}

		got := atomic.LoadInt32(&ct)
		want := int32(i)
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}

	sm.Stop()
}
