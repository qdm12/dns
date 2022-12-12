package services

type Service interface {
	Starter
	Stopper
	// String returns the service name.
	// It is assumed to be constant over the lifetime of the service.
	String() string
}

type Starter interface {
	// String returns the starter name.
	// It is assumed to be constant over the lifetime of the starter.
	String() string
	// Start starts the service.
	// On success, it returns a run error channel and a nil error.
	// On failure, it returns a nil run error channel and an error.
	// When the service is stopped, the service should close the run error channel.
	Start() (runError <-chan error, startErr error)
}

type Stopper interface {
	// String returns the stopper name.
	// It is assumed to be constant over the lifetime of the stopper.
	String() string
	// Stops stops the service.
	// A service should NOT write an error to its run error channel
	// if it is stopped.
	Stop() (err error)
}
