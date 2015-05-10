package sigmon_test

import (
	"fmt"
	"time"

	"github.com/codemodus/sigmon"
)

var (
	myServer = &server{}
)

type server struct {
	sm *sigmon.SignalMonitor
}

func (s *server) signalHandler()           {}
func (s *server) reload(withReconfig bool) {}
func (s *server) stop()                    {}

func ExampleSignalMonitor() {
	main := func() {
		myServer.sm = sigmon.New(nil)
		myServer.sm.Run()
		// Do things which cannot be affected by OS signals...

		myServer.sm.Set(myServer.signalHandler)
		// Do things which can be affected by OS signals...

		myServer.sm.Set(nil)
		// Do more things which cannot be affected by OS signals...

		myServer.sm.Stop()
		// OS signals will be handled normally.
	}
	main()
}

func ExampleSignalMonitor_signalHandlerFunc() {
	// contents of a func (s *server) signalHandler() {}
	// formatted to satisfy go test
	myServer.sm = sigmon.New(nil)
	signalHandler := func() {
		switch myServer.sm.Sig() {
		case sigmon.SIGHUP:
			myServer.reload(false)
		case sigmon.SIGINT, sigmon.SIGTERM:
			myServer.stop()
		case sigmon.SIGUSR1:
			myServer.reload(true)
		case sigmon.SIGUSR2:
			fmt.Println("USR2")
		}
	}
	signalHandler()

	// contents of a func (s *server) reload(withReconfig bool) {}
	// formatted to satisfy go test
	reload := func(withReconfig bool) {
		t1 := time.Now()
		if withReconfig {
			// Reload config.
		}
		t2 := time.Now()
		fmt.Println(myServer.sm.Sig(), t2.Sub(t1))
		// Output example: "HUP 156.78µs"

	}
	reload(true)
}
