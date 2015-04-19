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
	mu         sync.Mutex
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
	t1 := time.Now()
	mu.Lock()
	// Reload config
	mu.Unlock()
	t2 := time.Now()
	fmt.Println(sm.GetLast(), t2.Sub(t1))
	// Output: HUP 156.78µs
}

func ExampleSignalMonitor_stopFunc() {
	// Only handle TERM signals.
	if sm.GetLast() == "TERM" {
		fmt.Println(sm.GetLast())
	}
	// Output: TERM
}
