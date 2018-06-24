package sigmon

import (
	"sync"
	"syscall"
)

// HandlerFunc is used to communicate system signal handling behavior.
type HandlerFunc func(*State)

// Signal represents handled system signals.
type Signal syscall.Signal

func (s Signal) String() string {
	switch syscall.Signal(s) {
	case syscall.SIGHUP:
		return "HUP"
	case syscall.SIGINT:
		return "INT"
	case syscall.SIGTERM:
		return "TERM"
	case syscall.SIGUSR1:
		return "USR1"
	case syscall.SIGUSR2:
		return "USR2"
	default:
		return "UNKNOWN"
	}
}

// Signal constants define handled system signals.
const (
	SIGHUP  = Signal(syscall.SIGHUP)
	SIGINT  = Signal(syscall.SIGINT)
	SIGTERM = Signal(syscall.SIGTERM)
	SIGUSR1 = Signal(syscall.SIGUSR1)
	SIGUSR2 = Signal(syscall.SIGUSR2)
)

// SignalMonitor helps manage system signal handling.
type SignalMonitor struct {
	sync.Mutex
	isOn bool

	done chan struct{}

	j *junction
	r *registry
}

// New allocates a new SignalMonitor and related helper types, and returns the
// SignalMonitor. When a nil arg is provided, no action will be taken during
// signal handling. Start must be called in order to begin intercepting or
// ignoring system signals.
func New(fn HandlerFunc) *SignalMonitor {
	return &SignalMonitor{
		done: make(chan struct{}, 1),
		j:    newJunction(),
		r:    newRegistry(fn),
	}
}

// Set stores a HandlerFunc to be called when a system signal is received. If a
// nil arg is provided, no action will be taken during signal handling. Only
// the most recently set HandlerFunc will be used.
func (m *SignalMonitor) Set(fn HandlerFunc) {
	m.r.loadBuffer(fn)
}

// preScan is a sub event loop used to ensure that system signal handling
// stoppage and redefinition are prioritized over signal checking. preScan
// returns false if the outer event loop management should collapse.
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

// scan is the primary event loop. It returns false if the outer event loop
// management should collapse.
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

// monitor orchestrates the accounting and management of the event loops.
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

// Start establishes system signal awareness.
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

// Stop disestablishes system signal awareness. Subsequently called system
// signals are handled normally.
func (m *SignalMonitor) Stop() {
	m.Lock()
	defer m.Unlock()

	if m.isOn {
		m.isOn = false
		m.done <- struct{}{}
	}
}
