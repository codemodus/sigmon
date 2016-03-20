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
    ctxWrap := &contextWrap{c: make(chan string), prefix: "called/wrapped - "}
    
    sm := sigmon.New(ctxWrap.prefixAndLowerCaseHandler)
    sm.Run()
    
    // Simulate system signal calls and print results.
    callOSSiganl(syscall.SIGINT)
    
    select {
    case result := <-ctxWrap.c:
        fmt.Println(result) // Output: "called/wrapped - int"
    case <-time.After(1 * time.Second):
        fmt.Println("timeout waiting for signal")
    }
    
    sm.Stop()
}
```

## More Info

N/A

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/sigmon)

## Benchmarks

N/A
