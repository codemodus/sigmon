package sigmon

import (
	"syscall"
	"time"
)

var (
	sigs = []syscall.Signal{
		syscall.SIGTERM,
	}
)

func sendOnAll(m *SignalMonitor) {
	m.off <- struct{}{}
	m.handler.registry <- func(sm *SignalMonitor) {}
	m.junction.sigterm <- syscall.SIGTERM
}

func sendOnAllCount() int {
	return 3
}

func receiveOnAll(j *signalJunction) bool {
	for i := 0; i < 1; i++ {
		select {
		case <-j.sigterm:
		case <-time.After(time.Microsecond * 100):
			return false
		}
	}

	return true
}
