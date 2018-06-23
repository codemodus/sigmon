package sigmon_test

import (
	"github.com/codemodus/sigmon"
)

func Example() {
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

func Example_detailed() {
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

func handle(*sigmon.State) {}

var creds = ""

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
