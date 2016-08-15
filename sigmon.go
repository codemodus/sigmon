// Package sigmon simplifies os.Signal handling.
package sigmon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Signal wraps the string type to reduce confusion when checking Sig.
type Signal string

// Signal constants are string representations of the handled os.Signals.
const (
	SIGHUP  Signal = "HUP"
	SIGINT  Signal = "INT"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)

type signalHandler struct {
	sync.Mutex
	h   func(*SignalMonitor)
	set chan func(*SignalMonitor)
}

func (s *signalHandler) setHandler(handler func(*SignalMonitor)) {
	s.Lock()
	defer s.Unlock()

	s.h = handler
}

func (s *signalHandler) handle(sm *SignalMonitor) {
	s.Lock()
	defer s.Unlock()

	if s.h != nil {
		s.h(sm)
	}
}

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	sync.Mutex
	sig Signal
	on  bool
	off chan struct{}

	handler *signalHandler
}

// New takes a function and returns a SignalMonitor.  When a nil arg is
// provided, no action will be taken during signal handling.  Run must be
// called in order to begin monitoring.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	return &SignalMonitor{
		off: make(chan struct{}, 1),
		handler: &signalHandler{
			h:   handler,
			set: make(chan func(*SignalMonitor), 1),
		},
	}
}

// Set allows the handler function to be added or removed.  Only the most
// recently passed function will have any relevance.
func (s *SignalMonitor) Set(handler func(*SignalMonitor)) {
	select {
	case <-s.handler.set:
	default:
	}

	s.handler.set <- handler
}

// Run starts signal monitoring.  If no function has been provided, no action
// will be taken during signal handling.  The os.Signal which was called will
// be stored as a typed string (Signal) within the SignalMonitor for retrieval
// using Sig. Stop should be called within the provided handler functions and
// is not a default behavior.
func (s *SignalMonitor) Run() {
	s.Lock()
	defer s.Unlock()

	if s.on {
		return
	}
	s.on = true

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go s.process(wg)

	wg.Wait()
}

func (s *SignalMonitor) process(wg *sync.WaitGroup) {
	h := make(chan os.Signal, 1)
	i := make(chan os.Signal, 1)
	t := make(chan os.Signal, 1)
	u1 := make(chan os.Signal, 1)
	u2 := make(chan os.Signal, 1)

	signal.Notify(h, syscall.SIGHUP)
	signal.Notify(i, syscall.SIGINT)
	signal.Notify(t, syscall.SIGTERM)

	notifyUSR(u1, u2)

	defer s.closeChan(h)
	defer s.closeChan(i)
	defer s.closeChan(t)
	defer s.closeChan(u1)
	defer s.closeChan(u2)

	wg.Done()

	for {
		s.monitorWithPriority(h, i, t, u1, u2)
	}
}

func (s *SignalMonitor) monitorWithPriority(h, i, t, u1, u2 chan os.Signal) {
	select {
	case <-s.off:
		return
	case fn := <-s.handler.set:
		s.handler.setHandler(fn)
	default:
		s.monitorWithoutPriority(h, i, t, u1, u2)
	}
}

func (s *SignalMonitor) monitorWithoutPriority(h, i, t, u1, u2 chan os.Signal) {
	select {
	case <-s.off:
		return
	case fn := <-s.handler.set:
		s.handler.setHandler(fn)
	case <-h:
		s.handle(SIGHUP)
	case <-i:
		s.handle(SIGINT)
	case <-t:
		s.handle(SIGTERM)
	case <-u1:
		s.handle(SIGUSR1)
	case <-u2:
		s.handle(SIGUSR2)
	}
}

// Stop ends the goroutine which monitors signals.
func (s *SignalMonitor) Stop() {
	s.Lock()
	defer s.Unlock()

	if s.on {
		s.on = false
		s.off <- struct{}{}
	}
}

func (s *SignalMonitor) setSig(sig Signal) {
	s.Lock()
	defer s.Unlock()

	s.sig = sig
}

// Sig returns a typed string (Signal) representing the most recently called
// os.Signal.
func (s *SignalMonitor) Sig() Signal {
	s.Lock()
	defer s.Unlock()

	return s.sig
}

func (s *SignalMonitor) handle(sig Signal) {
	s.setSig(sig)
	s.handler.handle(s)
}

func (s *SignalMonitor) closeChan(c chan os.Signal) {
	signal.Stop(c)
	close(c)
}
