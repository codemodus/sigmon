// Package sigmon helps in managing HUP, INT, and TERM os.Signal behavior
// within an application.
package sigmon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SignalMonitor holds and calls funcs when called by relevant signals.
type SignalMonitor struct {
	mx        sync.RWMutex
	rFunc     func()
	sFunc     func()
	sig       string
	isRunning bool
	kill      chan bool
}

// New takes a reload and stop function and returns a set SignalMonitor.
// When a nil arg is provided, no action will be taken during the relevant
// signal(s).  Run must be called in order to begin monitoring.
func New(reloadFunc, stopFunc func()) (s *SignalMonitor) {
	s = &SignalMonitor{rFunc: reloadFunc, sFunc: stopFunc}
	return s
}

// Set allows functions to be added or removed to increase flexibility.
func (sm *SignalMonitor) Set(reloadFunc, stopFunc func()) {
	sm.mx.Lock()
	sm.rFunc = reloadFunc
	sm.sFunc = stopFunc
	sm.mx.Unlock()
}

// Run starts signal monitoring.  If functions have been provided, they will
// be called during the relevant case.  The os.Signal which was called will
// also be stored as a string within the SignalMonitor for retrieval (GetLast).
func (sm *SignalMonitor) Run() {
	if !sm.isRunning {
		sm.kill = make(chan bool)
		go func(s *SignalMonitor) {
			h := make(chan os.Signal, 1)
			i := make(chan os.Signal, 1)
			t := make(chan os.Signal, 1)
			signal.Notify(h, syscall.SIGHUP)
			signal.Notify(i, syscall.SIGINT)
			signal.Notify(t, syscall.SIGTERM)
			for {
				select {
				case <-h:
					s.sig = "HUP"
					if s.rFunc != nil {
						s.mx.RLock()
						s.rFunc()
						s.mx.RUnlock()
					}
				case <-i:
					s.sig = "INT"
					if s.sFunc != nil {
						s.mx.RLock()
						s.sFunc()
						s.mx.RUnlock()
						s.isRunning = false
						return
					}
				case <-t:
					s.sig = "TERM"
					if s.sFunc != nil {
						s.mx.RLock()
						s.sFunc()
						s.mx.RUnlock()
						s.isRunning = false
						return
					}
				case <-s.kill:
					s.isRunning = false
					return
				}
			}
		}(sm)
	}
}

// Stop kills the goroutine which is monitoring signals.
func (sm *SignalMonitor) Stop() {
	if sm.isRunning && sm.kill != nil {
		sm.kill <- true
	}
}

// GetLast returns a string of the most recently called os.Signal.
func (sm *SignalMonitor) GetLast() string {
	return sm.sig
}
