// Package sigmon simplifies os.Signal handling.
package sigmon

import (
	"sync"
)

// HandlerFunc ...
type HandlerFunc func(*State)

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

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	sync.Mutex
	isOn bool

	done chan struct{}

	j *junction
	r *registry
}

// New takes a function and returns a SignalMonitor. When a nil arg is
// provided, no action will be taken during signal handling. Start must be
// called in order to begin handling.
func New(fn HandlerFunc) *SignalMonitor {
	return &SignalMonitor{
		done: make(chan struct{}, 1),
		j:    newJunction(),
		r:    newRegistry(fn),
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
		m.r.get()(newState(s))
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
