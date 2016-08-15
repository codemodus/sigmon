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

type signalJunction struct {
	sync.Mutex
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

func (s *signalJunction) connect() {
	s.Lock()
	defer s.Unlock()

	signal.Notify(s.sighup, syscall.SIGHUP)
	signal.Notify(s.sigint, syscall.SIGINT)
	signal.Notify(s.sigterm, syscall.SIGTERM)
	notifyUSR(s.sigusr1, s.sigusr2)
}

func (s *signalJunction) disconnect() {
	defer signal.Stop(s.sighup)
	defer signal.Stop(s.sigint)
	defer signal.Stop(s.sigterm)
	defer signal.Stop(s.sigusr1)
	defer signal.Stop(s.sigusr2)
}

type signalHandler struct {
	sync.Mutex
	handler  func(*SignalMonitor)
	registry chan func(*SignalMonitor)
}

func (s *signalHandler) register(handler func(*SignalMonitor)) {
	select {
	case <-s.registry:
	default:
	}

	s.registry <- handler
}

func (s *signalHandler) set(handler func(*SignalMonitor)) {
	s.Lock()
	defer s.Unlock()

	s.handler = handler
}

func (s *signalHandler) handle(sm *SignalMonitor) {
	s.Lock()
	defer s.Unlock()

	if s.handler != nil {
		s.handler(sm)
	}
}

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	sync.Mutex
	sig Signal
	on  bool
	off chan struct{}

	junction *signalJunction
	handler  *signalHandler
}

// New takes a function and returns a SignalMonitor.  When a nil arg is
// provided, no action will be taken during signal handling.  Run must be
// called in order to begin monitoring.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	return &SignalMonitor{
		off:      make(chan struct{}, 1),
		junction: newSignalJunction(),
		handler: &signalHandler{
			handler:  handler,
			registry: make(chan func(*SignalMonitor), 1),
		},
	}
}

// Set allows the handler function to be added or removed.  Only the most
// recently passed function will have any relevance.
func (s *SignalMonitor) Set(handler func(*SignalMonitor)) {
	s.handler.register(handler)
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

	go s.monitor(wg)

	wg.Wait()
}

func (s *SignalMonitor) monitor(wg *sync.WaitGroup) {
	s.junction.connect()
	defer s.junction.disconnect()

	wg.Done()

	for {
		s.preScan()
	}
}

func (s *SignalMonitor) preScan() {
	select {
	case <-s.off:
		return
	case fn := <-s.handler.registry:
		s.handler.set(fn)
	default:
		s.scan()
	}
}

func (s *SignalMonitor) scan() {
	select {
	case <-s.off:
		return
	case fn := <-s.handler.registry:
		s.handler.set(fn)
	case <-s.junction.sighup:
		s.handle(SIGHUP)
	case <-s.junction.sigint:
		s.handle(SIGINT)
	case <-s.junction.sigterm:
		s.handle(SIGTERM)
	case <-s.junction.sigusr1:
		s.handle(SIGUSR1)
	case <-s.junction.sigusr2:
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
