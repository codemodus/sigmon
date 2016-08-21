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
	}
)

func sendOnAll(m *SignalMonitor) {
	m.off <- struct{}{}
	m.handler.registry <- func(sm *SignalMonitor) {}
	m.junction.sighup <- syscall.SIGHUP
	m.junction.sigint <- syscall.SIGINT
	m.junction.sigterm <- syscall.SIGTERM
}

func sendOnAllCount() int {
	return 5
}

func receiveOnAll(j *signalJunction) bool {
	for i := 0; i < 3; i++ {
		select {
		case <-j.sighup:
		case <-j.sigint:
		case <-j.sigterm:
		case <-time.After(time.Microsecond * 100):
			return false
		}
	}

	return true
}
