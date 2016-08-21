// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon_test

import (
	"fmt"
	"os"
	"sync"
	"syscall"

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

func Example_passingContext() {
	sigCtx := &signalContext{id: 123}

	// The setOutput method is ran on any signal and will store the signal text.
	sm := sigmon.New(sigCtx.setOutput)
	sm.Run()

	// Simulate system signal call (windows does not support self-signalling).
	if err := callOSSignal(syscall.SIGINT); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	sm.Stop()

	// The output method returns the called signal text and sigCtx.id value.
	fmt.Println(sigCtx.output())
	// Output:
	// INT 123
}

func signalHandler(sm *sigmon.SignalMonitor) {
	// intentionally left empty
}

type signalContext struct {
	sync.Mutex
	id  int
	out string
}

func (ctx *signalContext) setOutput(sm *sigmon.SignalMonitor) {
	ctx.Lock()
	defer ctx.Unlock()

	ctx.out = fmt.Sprintf("%s %d", sm.Sig(), ctx.id)
}

func (ctx *signalContext) output() string {
	ctx.Lock()
	defer ctx.Unlock()

	return ctx.out
}
