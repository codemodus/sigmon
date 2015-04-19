// Package sigmon helps in managing HUP, INT, and TERM os.Signal behavior
// within an application.
package sigmon

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalMonitor holds and calls funcs when called by relevant signals.
type SignalMonitor struct {
	reload func()
	stop   func()
	sig    string
	isOn   bool
	off    chan bool
	set    chan sigFuncs
}

type sigFuncs struct {
	reload func()
	stop   func()
}

// New takes a reload and stop function and returns a set SignalMonitor.
// When a nil arg is provided, no action will be taken during the relevant
// signal.  Run must be called in order to begin monitoring.
func New(reload, stop func()) (s *SignalMonitor) {
	s = &SignalMonitor{reload: reload, stop: stop,
		off: make(chan bool), set: make(chan sigFuncs, 1)}
	return s
}

// Set allows functions to be added or removed.  Only the most recently passed
// functions will have any relevance.
func (sm *SignalMonitor) Set(reload, stop func()) {
	select {
	case <-sm.set:
	default:
	}
	sm.set <- sigFuncs{reload: reload, stop: stop}
}

// Run starts signal monitoring.  If functions have been provided, they will
// be called during the relevant case.  The os.Signal which was called will
// also be stored as a string within the SignalMonitor for retrieval using
// GetLast.  Stop should be called within the provided functions and is not
// a default behavior of either INT or TERM.
func (sm *SignalMonitor) Run() {
	if !sm.isOn {
		go func(s *SignalMonitor) {
			h := make(chan os.Signal, 1)
			i := make(chan os.Signal, 1)
			t := make(chan os.Signal, 1)
			signal.Notify(h, syscall.SIGHUP)
			signal.Notify(i, syscall.SIGINT)
			signal.Notify(t, syscall.SIGTERM)

			for {
				select {
				case fns := <-s.set:
					s.reload = fns.reload
					s.stop = fns.stop
				case <-h:
					s.sig = "HUP"
					if s.reload != nil {
						s.reload()
					}
				case <-i:
					s.sig = "INT"
					if s.stop != nil {
						s.stop()
					}
				case <-t:
					s.sig = "TERM"
					if s.stop != nil {
						s.stop()
					}
				case <-s.off:
					s.isOn = false
					return
				}
			}
		}(sm)
	}
}

// Stop kills the goroutine which is monitoring signals.
func (sm *SignalMonitor) Stop() {
	if sm.isOn {
		sm.off <- true
	}
}

// GetLast returns a string of the most recently called os.Signal.
func (sm *SignalMonitor) GetLast() string {
	return sm.sig
}
