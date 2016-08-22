# sigmon

    go get github.com/codemodus/sigmon

Package sigmon simplifies os.Signal handling.

## Usage

```go
type Signal
type SignalMonitor
    func New(handler func(*SignalMonitor)) (s *SignalMonitor)
    func (s *SignalMonitor) Run()
    func (s *SignalMonitor) Set(handler func(*SignalMonitor))
    func (s *SignalMonitor) Sig() Signal
    func (s *SignalMonitor) Stop()
```

### Setup

```go
import (
    "github.com/codemodus/sigmon"
)

func main() {
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
```

### Signal Handler

```go
func signalHandler(sm *sigmon.SignalMonitor) {
    switch sm.Sig() {
    case sigmon.SIGHUP:
        // Reload
    case sigmon.SIGINT, sigmon.SIGTERM:
        // Stop
    case sigmon.SIGUSR1, sigmon.SIGUSR2:
        // More
    }
}
```

### Signal Handler With Context

```go
func main() {
    sigCtx := &signalContext{id: 123}

    // The setOutput method is ran on any signal and will store the signal text.
    sm := sigmon.New(sigCtx.setOutput)
    sm.Run()

    // Simulate system signal call (windows does not support self-signaling).
    if err := callOSSignal(syscall.SIGINT); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    sm.Stop()

    // The output method returns the called signal text and sigCtx.id value.
    fmt.Println(sigCtx.output()) // Outputs: "INT 123"
}
```

## More Info

### Windows Compatibility

sigmon will run on Windows systems without error. In order for this to be, 
notifications of USR1 and USR2 signals are detented as they are not supported 
whatsoever in Windows. All tests work on \*nix systems, but are not run on 
Windows. It is up to the user to assess whether their application is receiving 
INT, TERM, and, HUP signals properly along with what that may mean for the 
design of the affected system. 

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/sigmon)

## Benchmarks

N/A
