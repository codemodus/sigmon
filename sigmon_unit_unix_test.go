// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon

import (
	"syscall"
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

func sendOnAll(m *SignalMonitor) {
	m.off <- struct{}{}
	m.handler.registry <- func(sm *SignalMonitor) {}
	m.junction.sighup <- syscall.SIGHUP
	m.junction.sigint <- syscall.SIGINT
	m.junction.sigterm <- syscall.SIGTERM
	m.junction.sigusr1 <- syscall.SIGUSR1
	m.junction.sigusr2 <- syscall.SIGUSR2
}

func sendOnAllCount() int {
	return 7
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
