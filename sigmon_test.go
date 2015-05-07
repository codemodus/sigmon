package sigmon_test

import (
	"fmt"
	"sync"
	"time"
	"log"

	"github.com/codemodus/sigmon"
)

var (
	reloadFunc func()
	stopFunc   func()
	mx         sync.Mutex
	sm         *sigmon.SignalMonitor
)

func ExampleSignalMonitor() {
	sm := sigmon.New(nil)
	sm.Run()
	// Do things which cannot be affected by OS signals...

	sm.Set(myServer.signalHandler)
	// Do things which can be affected by OS signals...

	sm.Set(nil)
	// Do more things which cannot be affected by OS signals...

	sm.Stop()
	// OS signals will be handled normally.
}

func ExampleSignalMonitor_signalHandlerFunc() {
	switch sm.Sig() {
	case sigmon.SIGHUP:
		myServer.reload(false)
	case sigmon.SIGINT, sigmon.SIGTERM:
		myServer.stop()
	case sigmon.SIGUSR1:
		myServer.reload(true)
	case sigmon.SIGUSR2:
		log.Println("USR2")
		// do nothing
	}

	// ... within the reload function.
	t1 := time.Now()
	myServer.Lock()
	// Reload config.
	myServer.Unlock()
	t2 := time.Now()
	fmt.Println(sm.GetLast(), t2.Sub(t1))
	// Output: HUP 156.78µs
}
