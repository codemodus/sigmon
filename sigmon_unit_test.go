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

func TestUnitNewSignalJunction(t *testing.T) {
	j := newSignalJunction()
	if nil == j {
		t.Fatal("wanted new signalJunction, got nil")
	}

	go func() {
		j.sighup <- syscall.SIGHUP
		j.sigint <- syscall.SIGINT
		j.sigterm <- syscall.SIGTERM
		j.sigusr1 <- syscall.SIGUSR1
		j.sigusr2 <- syscall.SIGUSR2
	}()

	if !receiveOnAll(j) {
		t.Fatal("should not wait forever")
	}
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
