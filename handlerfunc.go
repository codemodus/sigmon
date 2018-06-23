package sigmon

import "sync"

// HandlerFunc ...
type HandlerFunc func(*State)

// handlerFuncRegistry is a support type for signalMonitor.
type handlerFuncRegistry struct {
	sync.Mutex
	mt  HandlerFunc
	fn  HandlerFunc
	reg chan HandlerFunc
}

func newHandlerFuncRegistry(fn HandlerFunc) *handlerFuncRegistry {
	mt := func(*State) {}

	return &handlerFuncRegistry{
		mt:  mt,
		fn:  filterHandlerFunc(mt, fn),
		reg: make(chan HandlerFunc, 1),
	}
}

func (r *handlerFuncRegistry) loadBuffer(fn HandlerFunc) {
	select {
	case <-r.reg:
	default:
	}

	r.reg <- r.filter(fn)
}

func (r *handlerFuncRegistry) buffer() chan HandlerFunc {
	return r.reg
}

func (r *handlerFuncRegistry) set(fn HandlerFunc) {
	r.Lock()
	defer r.Unlock()

	r.fn = r.filter(fn)
}

func (r *handlerFuncRegistry) get() HandlerFunc {
	r.Lock()
	defer r.Unlock()

	return r.filter(r.fn)
}

func (r *handlerFuncRegistry) filter(fn HandlerFunc) HandlerFunc {
	return filterHandlerFunc(r.mt, fn)
}

func filterHandlerFunc(mt, fn HandlerFunc) HandlerFunc {
	if fn != nil {
		return fn
	}

	if mt != nil {
		return mt
	}

	return func(*State) {}
}
