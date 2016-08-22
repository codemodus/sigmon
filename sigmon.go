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

// Signal constants are string representations of handled os.Signals.
const (
	SIGHUP  Signal = "HUP"
	SIGINT  Signal = "INT"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)

// signalJunction is a support type for signalMonitor.
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

func (j *signalJunction) connect() {
	j.Lock()
	defer j.Unlock()

	signal.Notify(j.sighup, syscall.SIGHUP)
	signal.Notify(j.sigint, syscall.SIGINT)
	signal.Notify(j.sigterm, syscall.SIGTERM)
	// split for unix/windows
	notifyUSR(j.sigusr1, j.sigusr2)
}

func (j *signalJunction) disconnect() {
	defer signal.Stop(j.sighup)
	defer signal.Stop(j.sigint)
	defer signal.Stop(j.sigterm)
	defer signal.Stop(j.sigusr1)
	defer signal.Stop(j.sigusr2)
}

// signalHandler is a support type for signalMonitor.
type signalHandler struct {
	sync.Mutex
	handler  func(*SignalMonitor)
	registry chan func(*SignalMonitor)
}

func newSignalHandler(handler func(*SignalMonitor)) *signalHandler {
	return &signalHandler{
		handler:  handler,
		registry: make(chan func(*SignalMonitor), 1),
	}
}

func (h *signalHandler) register(handler func(*SignalMonitor)) {
	select {
	case <-h.registry:
	default:
	}

	h.registry <- handler
}

func (h *signalHandler) set(handler func(*SignalMonitor)) {
	h.Lock()
	defer h.Unlock()

	h.handler = handler
}

func (h *signalHandler) handle(sm *SignalMonitor) {
	h.Lock()
	defer h.Unlock()

	if h.handler != nil {
		h.handler(sm)
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

// New takes a function and returns a SignalMonitor. When a nil arg is
// provided, no action will be taken during signal handling. Run must be
// called in order to begin handling.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	return &SignalMonitor{
		off:      make(chan struct{}, 1),
		junction: newSignalJunction(),
		handler:  newSignalHandler(handler),
	}
}

// Set allows the handler function to be added or removed. If no function has
// been provided, no action will be taken during signal handling. Only the most
// recently passed function holds any effect.
func (m *SignalMonitor) Set(handler func(*SignalMonitor)) {
	m.handler.register(handler)
}

func (m *SignalMonitor) scan() (alive bool) {
	select {
	case <-m.off:
		return false
	case fn := <-m.handler.registry:
		m.handler.set(fn)
	case <-m.junction.sighup:
		m.handle(SIGHUP)
	case <-m.junction.sigint:
		m.handle(SIGINT)
	case <-m.junction.sigterm:
		m.handle(SIGTERM)
	case <-m.junction.sigusr1:
		m.handle(SIGUSR1)
	case <-m.junction.sigusr2:
		m.handle(SIGUSR2)
	}

	return true
}

func (m *SignalMonitor) biasedScan() (alive bool) {
	select {
	case <-m.off:
		return false
	case fn := <-m.handler.registry:
		m.handler.set(fn)
	default:
		return m.scan()
	}

	return true
}

func (m *SignalMonitor) monitor(wg *sync.WaitGroup) {
	m.junction.connect()
	defer m.junction.disconnect()

	wg.Done()

	for {
		if !m.biasedScan() {
			return
		}
	}
}

// Run starts signal handling.
func (m *SignalMonitor) Run() {
	m.Lock()
	defer m.Unlock()

	if m.on {
		return
	}
	m.on = true

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go m.monitor(wg)

	wg.Wait()
}

// Stop discontinues all os.Signal handling.
func (m *SignalMonitor) Stop() {
	m.Lock()
	defer m.Unlock()

	if m.on {
		m.on = false
		m.off <- struct{}{}
	}
}

// Sig returns a typed string (Signal) representing the most recently called
// os.Signal.
func (m *SignalMonitor) Sig() Signal {
	m.Lock()
	defer m.Unlock()

	return m.sig
}

func (m *SignalMonitor) setSig(sig Signal) {
	m.Lock()
	defer m.Unlock()

	m.sig = sig
}

func (m *SignalMonitor) handle(sig Signal) {
	m.setSig(sig)
	m.handler.handle(m)
}
