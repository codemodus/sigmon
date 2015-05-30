// Package sigmon simplifies os.Signal handling.
package sigmon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

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
	handler func(*SignalMonitor)
	sig     Signal
	isOn    bool
	off     chan bool
	set     chan func(*SignalMonitor)
}

// Signal wraps the string type to reduce confusion when checking Sig.
type Signal string

// New takes a function and returns a SignalMonitor.  When a nil arg is
// provided, no action will be taken during signal handling.  Run must be
// called in order to begin monitoring.
func New(handler func(*SignalMonitor)) (s *SignalMonitor) {
	s = &SignalMonitor{handler: handler, off: make(chan bool),
		set: make(chan func(*SignalMonitor), 1),
	}
	return s
}

// Set allows the handler function to be added or removed.  Only the most
// recently passed function will have any relevance.
func (s *SignalMonitor) Set(handler func(*SignalMonitor)) {
	select {
	case <-s.set:
	default:
	}
	s.set <- handler
}

// Run starts signal monitoring.  If no function has been provided, no action
// will be taken during signal handling.  The os.Signal which was called will
// be stored as a string within the SignalMonitor for retrieval using GetLast.
// Stop should be called within the provided functions and is not a default
// behavior.
func (s *SignalMonitor) Run() {
	if !s.isOn {
		s.isOn = true
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(sm *SignalMonitor) {
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
			wg.Done()

			for {
				select {
				case <-sm.off:
					return
				case f := <-sm.set:
					sm.handler = f
				case <-h:
					sm.sig = SIGHUP
					if sm.handler != nil {
						sm.handler(sm)
					}
				case <-i:
					sm.sig = SIGINT
					if sm.handler != nil {
						sm.handler(sm)
					}
				case <-t:
					sm.sig = SIGTERM
					if sm.handler != nil {
						sm.handler(sm)
					}
				case <-u1:
					sm.sig = SIGUSR1
					if sm.handler != nil {
						sm.handler(sm)
					}
				case <-u2:
					sm.sig = SIGUSR2
					if sm.handler != nil {
						sm.handler(sm)
					}
				}
			}
		}(s)
		wg.Wait()
	}
}

// Stop kills the goroutine which is monitoring signals.
func (s *SignalMonitor) Stop() {
	if s.isOn {
		s.off <- true
		s.isOn = false
	}
}

// Sig returns a string of the most recently called os.Signal.
func (s *SignalMonitor) Sig() Signal {
	return s.sig
}
