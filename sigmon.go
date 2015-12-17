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

// {Signal} constants are text representations of the handled os.Signals.
const (
	SIGHUP  Signal = "HUP"
	SIGINT  Signal = "INT"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)

// SignalMonitor helps manage signal handling.
type SignalMonitor struct {
	hmu     sync.Mutex
	handler func(*SignalMonitor)

	mu   sync.Mutex
	sig  Signal
	isOn bool

	offc chan struct{}
	setc chan func(*SignalMonitor)
}

// New takes a function and returns a SignalMonitor.  When a nil arg is
// provided, no action will be taken during signal handling.  Run must be
// called in order to begin monitoring.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	s = &SignalMonitor{
		handler: handler,
		offc:    make(chan struct{}, 1),
		setc:    make(chan func(*SignalMonitor), 1),
	}

	return s
}

// Set allows the handler function to be added or removed.  Only the most
// recently passed function will have any relevance.
func (s *SignalMonitor) Set(handler func(*SignalMonitor)) {
	select {
	case <-s.setc:
	default:
	}

	s.setc <- handler
}

func (s *SignalMonitor) setHandler(handler func(*SignalMonitor)) {
	s.hmu.Lock()
	defer s.hmu.Unlock()

	s.handler = handler
}

// Run starts signal monitoring.  If no function has been provided, no action
// will be taken during signal handling.  The os.Signal which was called will
// be stored as a string within the SignalMonitor for retrieval using Sig.
// Stop should be called within the provided functions and is not a default
// behavior.
func (s *SignalMonitor) Run() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOn {
		return
	}
	s.isOn = true

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
	signal.Notify(u1, syscall.SIGUSR1)
	signal.Notify(u2, syscall.SIGUSR2)

	defer s.closeChan(h)
	defer s.closeChan(i)
	defer s.closeChan(t)
	defer s.closeChan(u1)
	defer s.closeChan(u2)

	wg.Done()

	for {
		select {
		case <-s.offc:
			return
		case fn := <-s.setc:
			s.setHandler(fn)
		default:
			select {
			case <-s.offc:
				return
			case fn := <-s.setc:
				s.setHandler(fn)
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
	}
}

// Stop kills the goroutine which is monitoring signals.
func (s *SignalMonitor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isOn {
		s.isOn = false
		s.offc <- struct{}{}
	}
}

func (s *SignalMonitor) setSig(sig Signal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sig = sig
}

// Sig returns a string of the most recently called os.Signal.
func (s *SignalMonitor) Sig() Signal {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sig
}

func (s *SignalMonitor) handle(sig Signal) {
	s.setSig(sig)

	s.hmu.Lock()
	defer s.hmu.Unlock()

	if s.handler != nil {
		s.handler(s)
	}
}

func (s *SignalMonitor) closeChan(c chan os.Signal) {
	signal.Stop(c)
	close(c)
}
