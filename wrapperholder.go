package main

type WrapperHolder struct {
	ping_wrappers []PingWrapperInterface
}

func (w *WrapperHolder) InitHosts(hosts []string, options Options, transition_writer *TransitionWriter) {
	w.ping_wrappers = make([]PingWrapperInterface, len(hosts))
	for i, host := range hosts {
		w.ping_wrappers[i] = NewPingWrapper(host, options, transition_writer)
	}
}

func (w *WrapperHolder) CalcStats(timeout_threshold int64) {
	for _, wrapper := range w.ping_wrappers {
		wrapper.CalcStats(timeout_threshold)
	}
}

func (w *WrapperHolder) Start() {
	for _, ping_wrapper := range w.ping_wrappers {
		ping_wrapper.Start()
	}
}

func (w *WrapperHolder) Stop() {
	for _, ping_wrapper := range w.ping_wrappers {
		ping_wrapper.Stop()
	}
}
