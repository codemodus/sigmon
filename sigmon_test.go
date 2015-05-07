package sigmon_test

import (
	"fmt"
	"time"

	"github.com/codemodus/sigmon"
)

var (
	sm *sigmon.SignalMonitor
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
		fmt.Println("USR2")
	}

	// ... within the reload function.
	t1 := time.Now()
	myServer.Lock()
	// Reload config.
	myServer.Unlock()
	t2 := time.Now()
	fmt.Println(sm.Sig(), t2.Sub(t1))
	// Output: HUP 156.78µs
}
