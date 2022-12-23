package services

// state is the state of a service.
type state uint8

const (
	// stateStopped is the state of a service that is stopped.
	stateStopped state = iota
	// stateStarting is the state of a service that is starting.
	stateStarting
	// stateRunning is the state of a service that is running.
	stateRunning
	// stateStopping is the state of a service that is stopping.
	stateStopping
	// stateCrashed is the state of a service that has crashed.
	stateCrashed
)
