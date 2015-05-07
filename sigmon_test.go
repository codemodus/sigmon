package sigmon_test

import (
	"fmt"
	"time"

	"github.com/codemodus/sigmon"
)

func ExampleSignalMonitor() {
	myServer.sm := sigmon.New(nil)
	myServer.sm.Run()
	// Do things which cannot be affected by OS signals...

	myServer.sm.Set(myServer.signalHandler)
	// Do things which can be affected by OS signals...

	myServer.sm.Set(nil)
	// Do more things which cannot be affected by OS signals...

	myServer.sm.Stop()
	// OS signals will be handled normally.
}

func ExampleSignalMonitor_signalHandlerFunc() {
	// Within (ms *myServer) myServer.signalHandler().
	switch sm.Sig() {
	case sigmon.SIGHUP:
		ms.reload(false)
	case sigmon.SIGINT, sigmon.SIGTERM:
		ms.stop()
	case sigmon.SIGUSR1:
		ms.reload(true)
	case sigmon.SIGUSR2:
		fmt.Println("USR2")
	}

	// Within (ms *myServer) myServer.reload(reconfig bool).
	t1 := time.Now()
	if reconfig {
		// Reload config.
	}
	t2 := time.Now()
	fmt.Println(ms.sm.Sig(), t2.Sub(t1))
	// Output: HUP 156.78µs
}
