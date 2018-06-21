package sigmon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// signalJunction is a support type for signalMonitor.
type signalJunction struct {
	sync.Mutex
	isConnected bool

	sighup  chan os.Signal
	sigint  chan os.Signal
	sigterm chan os.Signal
	sigusr1 chan os.Signal
	sigusr2 chan os.Signal
}

func newSignalJunction() *signalJunction {
	return &signalJunction{
		sighup:  make(chan os.Signal, 1),
		sigint:  make(chan os.Signal, 1),
		sigterm: make(chan os.Signal, 1),
		sigusr1: make(chan os.Signal, 1),
		sigusr2: make(chan os.Signal, 1),
	}
}

func (j *signalJunction) connect() {
	j.Lock()
	defer j.Unlock()

	if j.isConnected {
		return
	}

	signal.Notify(j.sighup, syscall.SIGHUP)
	signal.Notify(j.sigint, syscall.SIGINT)
	signal.Notify(j.sigterm, syscall.SIGTERM)
	// split for unix/windows
	notifyUSR(j.sigusr1, j.sigusr2)

	j.isConnected = true
}

func (j *signalJunction) disconnect() {
	j.Lock()
	defer j.Unlock()

	if !j.isConnected {
		return
	}

	j.isConnected = false

	defer signal.Stop(j.sighup)
	defer signal.Stop(j.sigint)
	defer signal.Stop(j.sigterm)
	defer signal.Stop(j.sigusr1)
	defer signal.Stop(j.sigusr2)
}
