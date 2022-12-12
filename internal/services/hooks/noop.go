package hooks

type NoopHooks struct{}

func NewNoop() *NoopHooks {
	return &NoopHooks{}
}

func (h *NoopHooks) OnStart(service string)              {}
func (h *NoopHooks) OnStarted(service string, err error) {}
func (h *NoopHooks) OnStop(service string)               {}
func (h *NoopHooks) OnStopped(service string, err error) {}
func (h *NoopHooks) OnCrash(service string, err error)   {}
