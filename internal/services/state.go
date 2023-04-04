package services

import "fmt"

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

func (s State) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateCrashed:
		return "crashed"
	default:
		panic(fmt.Sprintf("State %d has not corresponding string", s))
	}
}
