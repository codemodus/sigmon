// Package sigmon simplifies os.Signal handling.
package sigmon

import (
	"sync"
)

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	sync.Mutex
	sig Signal
	on  bool

	done chan struct{}

	j *signalJunction
	h *signalHandler
}

// New takes a function and returns a SignalMonitor. When a nil arg is
// provided, no action will be taken during signal handling. Start must be
// called in order to begin handling.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	return &SignalMonitor{
		done: make(chan struct{}, 1),
		j:    newSignalJunction(),
		h:    newSignalHandler(handler),
	}
}

// Set allows the handler function to be added or removed. If no function has
// been provided, no action will be taken during signal handling. Only the most
// recently passed function holds any effect.
func (m *SignalMonitor) Set(handler func(*SignalMonitor)) {
	m.h.register(handler)
}

func (m *SignalMonitor) preScan() (alive bool) {
	select {
	case <-m.done:
		return false
	case fn := <-m.h.registry:
		m.h.set(fn)
	default:
	}

	return true
}

func (m *SignalMonitor) scan() (alive bool) {
	select {
	case <-m.done:
		return false
	case fn := <-m.h.registry:
		m.h.set(fn)
	case <-m.j.sighup:
		m.handle(SIGHUP)
	case <-m.j.sigint:
		m.handle(SIGINT)
	case <-m.j.sigterm:
		m.handle(SIGTERM)
	case <-m.j.sigusr1:
		m.handle(SIGUSR1)
	case <-m.j.sigusr2:
		m.handle(SIGUSR2)
	}

	return true
}

func (m *SignalMonitor) monitor(wg *sync.WaitGroup) {
	m.j.connect()
	defer m.j.disconnect()

	wg.Done()

	for {
		if !m.preScan() {
			return
		}

		if !m.scan() {
			return
		}
	}
}

// Start starts signal handling.
func (m *SignalMonitor) Start() {
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
		m.done <- struct{}{}
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
	m.h.handle(m)
}
