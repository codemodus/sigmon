package sigmon

import "sync"

// signalHandler is a support type for signalMonitor.
type signalHandler struct {
	sync.Mutex
	h func(*SignalMonitor)

	registry chan func(*SignalMonitor)
}

func newSignalHandler(handler func(*SignalMonitor)) *signalHandler {
	return &signalHandler{
		h:        handler,
		registry: make(chan func(*SignalMonitor), 1),
	}
}

func (h *signalHandler) register(handler func(*SignalMonitor)) {
	select {
	case <-h.registry:
	default:
	}

	h.registry <- handler
}

func (h *signalHandler) set(handler func(*SignalMonitor)) {
	h.Lock()
	defer h.Unlock()

	h.h = handler
}

func (h *signalHandler) handle(sm *SignalMonitor) {
	h.Lock()
	defer h.Unlock()

	if h.h != nil {
		h.h(sm)
	}
}
