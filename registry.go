package sigmon

import "sync"

// registry is a support type for signalMonitor.
type registry struct {
	sync.Mutex
	mt  HandlerFunc
	fn  HandlerFunc
	reg chan HandlerFunc
}

func newRegistry(fn HandlerFunc) *registry {
	mt := func(*State) {}

	return &registry{
		mt:  mt,
		fn:  filterFunc(mt, fn),
		reg: make(chan HandlerFunc, 1),
	}
}

func (r *registry) loadBuffer(fn HandlerFunc) {
	select {
	case <-r.reg:
	default:
	}

	r.reg <- r.filter(fn)
}

func (r *registry) buffer() chan HandlerFunc {
	return r.reg
}

func (r *registry) set(fn HandlerFunc) {
	r.Lock()
	defer r.Unlock()

	r.fn = r.filter(fn)
}

func (r *registry) get() HandlerFunc {
	r.Lock()
	defer r.Unlock()

	return r.filter(r.fn)
}

func (r *registry) filter(fn HandlerFunc) HandlerFunc {
	return filterFunc(r.mt, fn)
}

func filterFunc(mt, fn HandlerFunc) HandlerFunc {
	if fn != nil {
		return fn
	}

	if mt != nil {
		return mt
	}

	return func(*State) {}
}
