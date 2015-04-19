package sigmon

import (
	"fmt"
	"sync"
	"time"

	"github.com/codemodus/sigmon"
)

var (
	reloadFunc func()
	stopFunc   func()
	mx         sync.Mutex
	sm         *SignalMonitor
)

func ExampleSignalMonitor() {
	sm := sigmon.New(nil, nil)
	sm.Run()

	// Do things which cannot be affected by OS signals...

	sm.Set(reloadFunc, stopFunc)

	// Do things which can be affected by OS signals...

	sm.Stop()

	// OS signals will be handled normally.
}

func ExampleSignalMonitor_reloadFunc() {
	// ... within the provided reload function.
	t1 := time.Now()
	mx.Lock()
	// Reload config.
	mx.Unlock()
	t2 := time.Now()
	fmt.Println(sm.GetLast(), t2.Sub(t1))
	// Output: HUP 156.78µs
}

func ExampleSignalMonitor_stopFunc() {
	// ... within the provided stop function.
	// Handle TERM and ignore INT.
	if sm.GetLast() == "TERM" {
		fmt.Println(sm.GetLast())
		sm.Stop()
		// Stop application rest of application.
	}
	// Output: TERM
}
