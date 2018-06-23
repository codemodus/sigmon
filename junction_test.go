// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestJunction(t *testing.T) {
	t.Run("ConnectDisconnectSignals", tJunctionConnectDisconnectSignals)
}

func tJunctionConnectDisconnectSignals(t *testing.T) {
	j := newSignalJunction()

	j.connect()
	j.connect() // panics if incorrectly designed

	for _, s := range osSignals {
		if err := callOSSignal(s); err != nil {
			t.Errorf("unexpected error when calling %s: %s", s, err)
		}
	}

	if !isPickingUp(j.signals(), len(osSignals)) {
		t.Fatalf("should not block")
	}

	j.disconnect()
	j.disconnect() // panics if incorrectly designed

	s := syscall.SIGHUP
	c := make(chan os.Signal, 1)
	signal.Notify(c, s)

	if err := callOSSignal(s); err != nil {
		t.Errorf("unexpected error when calling %s: %s", s, err)
	}

	if isPickingUp(j.signals(), 1) {
		t.Fatalf("should block")
	}

	select {
	case <-c:
		signal.Stop(c)
	default:
		t.Fatalf("should not block")
	}
}

var (
	osSignals = []syscall.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	}
)

func callOSSignal(s syscall.Signal) error {
	if err := syscall.Kill(syscall.Getpid(), s); err != nil {
		return err
	}

	// delay for signal propagation
	for i := 1 << 23; i > 0; i-- {
	}

	return nil
}

func isPickingUp(c chan Signal, ct int) bool {
	for i := 0; i < ct; i++ {
		select {
		case <-c:
		case <-time.After(time.Microsecond * 100):
			return false
		}
	}

	return true
}
