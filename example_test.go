package sigmon_test

import (
	"syscall"

	"github.com/codemodus/sigmon/v2"
)

func Example() {
	sm := sigmon.New(nil)
	sm.Start()

	// Do things which cannot be affected by OS signals other than SIGKILL...

	sm.Set(handle)

	// Do things which can be affected by handled OS signals...

	sm.Stop()
	// OS signals will be handled normally.
}

func Example_elaborated() {
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

func Example_handlerFunc() {
	handle = func(s *sigmon.State) {
		switch s.Signal() {
		case sigmon.SIGHUP:
			server.Restart()
		default:
			server.Shutdown()
		}
	}
}

func Example_handlerFuncSyscall() {
	handle = func(s *sigmon.State) {
		switch syscall.Signal(s.Signal()) {
		case syscall.SIGHUP:
			server.Restart()
		default:
			server.Shutdown()
		}
	}
}

type dataBase struct {
	Migrate func()
}

func newDataBase(string) *dataBase {
	return &dataBase{
		Migrate: func() {},
	}
}

type webApp struct {
	ListenAndServe func()
	Restart        func()
	Shutdown       func()
	Wait           func()
}

func newWebApp(db *dataBase) *webApp {
	return &webApp{
		ListenAndServe: func() {},
		Restart:        func() {},
		Shutdown:       func() {},
		Wait:           func() {},
	}
}

var (
	creds  = ""
	server = newWebApp(nil)

	handle sigmon.HandlerFunc = func(*sigmon.State) {}
)
