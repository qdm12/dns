package hooks

type NoopHooks struct{}

func NewNoop() *NoopHooks {
	return &NoopHooks{}
}

func (h *NoopHooks) OnStart(string)          {}
func (h *NoopHooks) OnStarted(string, error) {}
func (h *NoopHooks) OnStop(string)           {}
func (h *NoopHooks) OnStopped(string, error) {}
func (h *NoopHooks) OnCrash(string, error)   {}
