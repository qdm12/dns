package services

// State is the State of a service.
type State uint8

const (
	// StateStopped is the state of a service that is stopped.
	StateStopped State = iota
	// StateStarting is the state of a service that is starting.
	StateStarting
	// StateRunning is the state of a service that is running.
	StateRunning
	// StateStopping is the state of a service that is stopping.
	StateStopping
	// StateCrashed is the state of a service that has crashed.
	StateCrashed
)
