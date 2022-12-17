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
	// If the service crashes, only one single error should be sent in
	// the error channel.
	// When the service is stopped, the service should NOT send an error
	// in the run error channel or close this one.
	Start() (runError <-chan error, startErr error)
}

type Stopper interface {
	// String returns the stopper name.
	// It is assumed to be constant over the lifetime of the stopper.
	String() string
	// Stops stops the service.
	// A service should NOT close or write an error to its run error channel
	// if it is stopped.
	Stop() (err error)
}
