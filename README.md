# sigmon

    go get -u github.com/codemodus/sigmon

Package sigmon simplifies os.Signal handling.

The benefits of this over a more simplistic approach are eased signal 
bypassability, eased signal handling replaceability, the rectification 
of system signal quirks, and operating system portability (Windows will 
ignore USR1 and USR2 signals, and some testing is bypassed).

## Usage

```go
type HandlerFunc
type Signal
type SignalMonitor
    func New(fn HandlerFunc) *SignalMonitor
    func (m *SignalMonitor) Set(fn HandlerFunc)
    func (m *SignalMonitor) Start()
    func (m *SignalMonitor) Stop()
type State
    func (s *State) Signal() Signal
```

### Setup

```go
import (
    "github.com/codemodus/sigmon/v2"
)

func main() {
    sm := sigmon.New(nil)
    sm.Start()

    // Do things which cannot be affected by OS signals other than SIGKILL...

    sm.Set(handle)

    // Do things which can be affected by handled OS signals...

    sm.Stop()
    // OS signals will be handled normally.
}
```

### HandlerFunc

```go
func handle(s *sigmon.State) {
    switch s.Signal() {
    case sigmon.SIGHUP:
        // reload
    case sigmon.SIGINT, sigmon.SIGTERM:
        // stop
    case sigmon.SIGUSR1, sigmon.SIGUSR2:
        // more
    }
}
```

### Setup Elaborated

```go
func main() {
    sm := sigmon.New(nil)
    sm.Start()

    db := newDataBase(creds)
    db.Migrate()

    app := newWebApp(db)
    app.ListenAndServe()

    sm.Set(func(s *sigmon.State) {
        switch s.Signal() {
        case sigmon.SIGHUP:
            app.Restart()
        default:
            app.Shutdown()
        }
    })

    app.Wait()
}
```

### HandlerFunc Using `syscall` Signal Constants

```go
func handle(s *sigmon.State) {
    switch syscall.Signal(s.Signal()) {
    case syscall.SIGHUP:
        // reload
    default:
        // stop
    }
}
```

## More Info

### Windows Compatibility

sigmon will run on Windows systems without error. In order for this to be, 
notifications of USR1 and USR2 signals are not wired up as they are not 
supported whatsoever in Windows. All tests work on \*nix systems, but are not 
run on Windows. It is up to the user to assess whether their application is 
receiving INT, TERM, and, HUP signals properly along with what that may mean 
for the design of the affected system. 

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/sigmon)

## Benchmarks

N/A
