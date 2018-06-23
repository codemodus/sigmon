package sigmon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// signalJunction manages the connection and disconnection of os signal
// handling.
type signalJunction struct {
	sync.Mutex
	isConnected bool

	sighup  chan os.Signal
	sigint  chan os.Signal
	sigterm chan os.Signal
	sigusr1 chan os.Signal
	sigusr2 chan os.Signal

	sigs chan Signal
	done chan struct{}
}

func newSignalJunction() *signalJunction {
	return &signalJunction{
		sighup:  make(chan os.Signal, 1),
		sigint:  make(chan os.Signal, 1),
		sigterm: make(chan os.Signal, 1),
		sigusr1: make(chan os.Signal, 1),
		sigusr2: make(chan os.Signal, 1),
		sigs:    make(chan Signal, 1),
		done:    make(chan struct{}, 1),
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
	notifyx(j.sigusr1, j.sigusr2) // split for unix/windows

	go func() {
		for {
			select {
			case <-j.done:
				return
			case <-j.sighup:
				j.sigs <- SIGHUP
			case <-j.sigint:
				j.sigs <- SIGINT
			case <-j.sigterm:
				j.sigs <- SIGTERM
			case <-j.sigusr1:
				j.sigs <- SIGUSR1
			case <-j.sigusr2:
				j.sigs <- SIGUSR2
			}
		}
	}()

	j.isConnected = true
}

func (j *signalJunction) disconnect() {
	j.Lock()
	defer j.Unlock()

	if !j.isConnected {
		return
	}

	j.done <- struct{}{}

	j.isConnected = false

	defer signal.Stop(j.sighup)
	defer signal.Stop(j.sigint)
	defer signal.Stop(j.sigterm)
	defer signal.Stop(j.sigusr1)
	defer signal.Stop(j.sigusr2)
}

func (j *signalJunction) signals() chan Signal {
	return j.sigs
}
