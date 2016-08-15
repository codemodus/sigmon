package sigmon_test

import (
	"fmt"
	"syscall"
	"time"

	"github.com/codemodus/sigmon"
)

type contextWrap struct {
	c      chan string
	prefix string
}

func Example() {
	sm := sigmon.New(nil)
	sm.Run()
	// Do things which cannot be affected by OS signals...

	sm.Set(signalHandler)
	// Do things which can be affected by OS signals...

	sm.Set(nil)
	// Do more things which cannot be affected by OS signals...

	sm.Stop()
	// OS signals will be handled normally.
}

func Example_signalHandlerFunc() {
	// ...

	signalHandler := func(sm *sigmon.SignalMonitor) {
		switch sm.Sig() {
		case sigmon.SIGHUP:
			sm.Set(nil)
			// Reload
			sm.Set(signalHandler)
		case sigmon.SIGINT, sigmon.SIGTERM:
			sm.Set(nil)
			// Stop
		case sigmon.SIGUSR1, sigmon.SIGUSR2:
			// More
		}
	}

	sm := sigmon.New(signalHandler)
	sm.Run()

	// ...
}

func Example_funWithContext() {
	ctxWrap := &contextWrap{
		c:      make(chan string),
		prefix: "called/wrapped - ",
	}

	sm := sigmon.New(ctxWrap.prefixAndLowerCaseHandler)
	sm.Run()

	// Simulate system signal calls and print results.
	callOSSiganl(syscall.SIGINT)

	select {
	case result := <-ctxWrap.c:
		fmt.Println(result)
	case <-time.After(time.Second):
		fmt.Println("timeout waiting for signal")
	}

	callOSSiganl(syscall.SIGHUP)

	select {
	case result := <-ctxWrap.c:
		fmt.Println(result)
	case <-time.After(time.Second):
		fmt.Println("timeout waiting for signal")
	}

	sm.Stop()

	// Output:
	// called/wrapped - int
	// called/wrapped - hup
}
