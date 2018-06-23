// Package sigmon simplifies os.Signal handling.
package sigmon

import (
	"sync"
)

// Signal wraps the string type to reduce confusion when checking Sig.
type Signal string

// Signal constants are string representations of handled os.Signals.
const (
	NOSIG   Signal = "N/A"
	SIGHUP  Signal = "HUP"
	SIGINT  Signal = "INT"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	sync.Mutex
	isOn bool
	s    *State

	done chan struct{}

	j *signalJunction
	r *handlerFuncRegistry
}

// New takes a function and returns a SignalMonitor. When a nil arg is
// provided, no action will be taken during signal handling. Start must be
// called in order to begin handling.
func New(fn HandlerFunc) *SignalMonitor {
	return &SignalMonitor{
		done: make(chan struct{}, 1),
		j:    newSignalJunction(),
		r:    newHandlerFuncRegistry(fn),
		s:    newState(NOSIG),
	}
}

// Set allows the handler function to be added or removed. If no function has
// been provided, no action will be taken during signal handling. Only the most
// recently passed function holds any effect.
func (m *SignalMonitor) Set(fn HandlerFunc) {
	m.r.loadBuffer(fn)
}

func (m *SignalMonitor) preScan() (alive bool) {
	select {
	case <-m.done:
		return false
	case fn := <-m.r.buffer():
		m.r.set(fn)
	default:
	}

	return true
}

func (m *SignalMonitor) scan() (alive bool) {
	select {
	case <-m.done:
		return false
	case fn := <-m.r.buffer():
		m.r.set(fn)
	case s := <-m.j.signals():
		m.setState(s)
		fn := m.r.get()
		fn(m.State())
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

	if m.isOn {
		return
	}
	m.isOn = true

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go m.monitor(wg)

	wg.Wait()
}

// Stop discontinues all os.Signal handling.
func (m *SignalMonitor) Stop() {
	m.Lock()
	defer m.Unlock()

	if m.isOn {
		m.isOn = false
		m.done <- struct{}{}
	}
}

// State returns a typed string (Signal) representing the most recently called
// os.Signal.
func (m *SignalMonitor) State() *State {
	m.Lock()
	defer m.Unlock()

	return m.s
}

func (m *SignalMonitor) setState(s Signal) {
	m.Lock()
	defer m.Unlock()

	m.s = newState(s)
}
