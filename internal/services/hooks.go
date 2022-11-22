package services

type Hooks interface {
	HooksStart
	HooksStop
	HooksCrash
}

type HooksStart interface {
	OnStart(service string)
	OnStarted(service string, err error)
}

type HooksStop interface {
	OnStop(service string)
	OnStopped(service string, err error)
}

type HooksCrash interface {
	OnCrash(service string, err error)
}
