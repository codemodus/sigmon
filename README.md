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
    "github.com/codemodus/sigmon"
)

func main() {
    sm := sigmon.New(nil)
    sm.Start()
    // Do things which cannot be affected by OS signals...

    sm.Set(handle)
    // Do things which can be affected by OS signals...

    sm.Set(nil)
    // Do more things which cannot be affected by OS signals...

    sm.Stop()
    // OS signals will be handled normally.
}
```

### Signal Handler

```go
func handle(s *sigmon.State) {
    switch s.Signal() {
    case sigmon.SIGHUP:
        // Reload
    case sigmon.SIGINT, sigmon.SIGTERM:
        // Stop
    case sigmon.SIGUSR1, sigmon.SIGUSR2:
        // More
    }
}
```

### Setup With More Details

```go
func main() {
    sm := sigmon.New(nil)
    sm.Start()
    // Only SIGKILL can disturb the following until sm.Set is called below.

    db := newDataBase(creds)
    db.Migrate()

    app := newWebApp(db)
    app.ListenAndServe()

    sm.Set(func(s *sigmon.State) {
        switch s.Signal() {
        case sigmon.SIGHUP:
            app.Restart()
        default:
            app.Shutdown() // shutdown on all other signals
        }
    })

    // Once app.Shutdown is called, app.Wait will stop blocking.
    app.Wait()
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
