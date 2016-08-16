package sigmon

import (
	"syscall"
	"testing"
	"time"
)

func TestUnitNewSignalJunction(t *testing.T) {
	j := newSignalJunction()
	if nil == j {
		t.Fatal("wanted new signalJunction, got nil")
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("panicked when accessing channel")
			}
		}()

		go func() {
			j.sighup <- syscall.SIGHUP
			j.sigint <- syscall.SIGHUP
			j.sigterm <- syscall.SIGHUP
			j.sigusr1 <- syscall.SIGHUP
			j.sigusr2 <- syscall.SIGHUP
		}()

		for i := 0; i < 5; i++ {
			select {
			case <-j.sighup:
			case <-j.sigint:
			case <-j.sigterm:
			case <-j.sigusr1:
			case <-j.sigusr2:
			case <-time.After(time.Second):
				t.Fatal("will likely wait forever")
			}
		}

	}()
}
