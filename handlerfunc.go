package sigmon

import "sync"

// HandlerFunc ...
type HandlerFunc func(*State)

// handlerFuncRegistry is a support type for signalMonitor.
type handlerFuncRegistry struct {
	sync.Mutex
	fn  HandlerFunc
	reg chan HandlerFunc
}

func newHandlerFuncRegistry(fn HandlerFunc) *handlerFuncRegistry {
	return &handlerFuncRegistry{
		fn:  fn,
		reg: make(chan HandlerFunc, 1),
	}
}

func (r *handlerFuncRegistry) load(fn HandlerFunc) {
	select {
	case <-r.reg:
	default:
	}

	r.reg <- fn
}

func (r *handlerFuncRegistry) crank(fn HandlerFunc) {
	r.Lock()
	defer r.Unlock()

	r.fn = fn
}

func (r *handlerFuncRegistry) handle(s *State) {
	r.Lock()
	defer r.Unlock()

	if r.fn != nil {
		r.fn(s)
	}
}
