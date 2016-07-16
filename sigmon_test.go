package sigmon_test

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
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

func TestSignalIgnorance(t *testing.T) {
	sm := sigmon.New(nil)
	sm.Run()

	testCallOSSiganl(t, syscall.SIGINT)

	sm.Stop()
}

func TestSignalHandling(t *testing.T) {
	cw := &contextWrap{c: make(chan string), prefix: "wrapped"}

	tests := []struct {
		h    func(*sigmon.SignalMonitor)
		send syscall.Signal
		recv string
	}{
		{cw.signalHandler, syscall.SIGHUP, string(sigmon.SIGHUP)},
		{cw.signalHandler, syscall.SIGINT, string(sigmon.SIGINT)},
		{cw.signalHandler, syscall.SIGTERM, string(sigmon.SIGTERM)},
		{cw.signalHandler, syscall.SIGUSR1, string(sigmon.SIGUSR1)},
		{cw.signalHandler, syscall.SIGUSR2, string(sigmon.SIGUSR2)},
		{cw.prefixAndLowerCaseHandler, syscall.SIGHUP, cw.prefix + "hup"},
		{cw.prefixAndLowerCaseHandler, syscall.SIGINT, cw.prefix + "int"},
		{cw.prefixAndLowerCaseHandler, syscall.SIGTERM, cw.prefix + "term"},
		{cw.prefixAndLowerCaseHandler, syscall.SIGUSR1, cw.prefix + "usr1"},
		{cw.prefixAndLowerCaseHandler, syscall.SIGUSR2, cw.prefix + "usr2"},
	}

	sm := sigmon.New(nil)
	sm.Run()

	testCallOSSiganl(t, syscall.SIGINT)

	for _, v := range tests {
		sm.Set(v.h)

		testCallOSSiganl(t, v.send)

		want := v.recv
		select {
		case got := <-cw.c:
			if got != want {
				t.Errorf("signal was %v, want %v", got, want)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("timeout waiting for %v", want)
		}
	}
}

func signalHandler(sm *sigmon.SignalMonitor) {
	switch sm.Sig() {
	case sigmon.SIGHUP, sigmon.SIGINT, sigmon.SIGTERM, sigmon.SIGUSR1, sigmon.SIGUSR2:
	}
}

func callOSSiganl(s syscall.Signal) {
	c := make(chan os.Signal)
	signal.Notify(c, s)
	if err := syscall.Kill(syscall.Getpid(), s); err != nil {
		fmt.Println(err)
	}
	select {
	case <-time.After(1 * time.Second):
		fmt.Printf("timeout waiting for %v", s)
	case <-c:
		// prevent syscall.Kill from "bleeding"
		time.Sleep(11 * time.Microsecond)
	}
	signal.Stop(c)
}

func (cw *contextWrap) signalHandler(sm *sigmon.SignalMonitor) {
	s := sm.Sig()
	switch s {
	case sigmon.SIGHUP, sigmon.SIGINT, sigmon.SIGTERM, sigmon.SIGUSR1, sigmon.SIGUSR2:
		cw.c <- string(s)
	}
}

func (cw *contextWrap) prefixAndLowerCaseHandler(sm *sigmon.SignalMonitor) {
	switch sm.Sig() {
	case sigmon.SIGHUP, sigmon.SIGINT, sigmon.SIGTERM, sigmon.SIGUSR1, sigmon.SIGUSR2:
		g := cw.contextHandler(sm)
		cw.c <- g
	}
}

func (cw *contextWrap) contextHandler(sm *sigmon.SignalMonitor) string {
	return cw.prefix + strings.ToLower(string(sm.Sig()))
}

func testCallOSSiganl(t *testing.T, s syscall.Signal) {
	c := make(chan os.Signal)
	signal.Notify(c, s)
	if err := syscall.Kill(syscall.Getpid(), s); err != nil {
		t.Fatal(err)
	}
	select {
	case <-c:
		// prevent syscall.Kill from "bleeding"
		time.Sleep(10 * time.Microsecond)
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %v", s)
	}
	signal.Stop(c)
}

func BenchmarkSimple(b *testing.B) {
	for n := 0; n < b.N; n++ {
		sm := sigmon.New(nil)
		sm.Run()
		sm.Stop()
	}
}
