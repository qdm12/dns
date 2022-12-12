package services

import (
	"fmt"
	"sync"
)

var _ Service = (*Sequence)(nil)

type Sequence struct {
	name          string
	running       bool
	servicesStart []Service
	servicesStop  []Service
	hooks         Hooks
	mutex         *sync.Mutex
	internalMutex *sync.Mutex
	fanIn         *errorsFanIn
	// runningServices contains service names that are currently running.
	runningServices map[string]struct{}
	interceptStop   chan struct{}
	interceptDone   chan struct{}
}

func NewSequence(settings SequenceSettings) (sequence *Sequence, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	servicesStart := make([]Service, len(settings.ServicesStart))
	copy(servicesStart, settings.ServicesStart)

	servicesStop := make([]Service, len(settings.ServicesStop))
	copy(servicesStop, settings.ServicesStop)

	return &Sequence{
		name:            settings.Name,
		servicesStart:   servicesStart,
		servicesStop:    servicesStop,
		hooks:           settings.Hooks,
		mutex:           &sync.Mutex{},
		internalMutex:   &sync.Mutex{},
		runningServices: make(map[string]struct{}, len(servicesStart)),
	}, nil
}

func (s *Sequence) String() string {
	if s.name == "" {
		return "sequence"
	}
	return "sequence " + s.name
}

// Start starts services in the order specified by the
// the sequence of services.
//
// If a service fails to start, the `startErr` is returned
// and all other running services are stopped in the order
// specified by the stop sequence of services.
//
// If a service fails after being started in this method call,
// the error is returned in `startErr` and all other running
// services are stopped in the order specified by the stop
// sequence of services.
//
// If a service fails after this method call returns without error,
// all other running services are stopped and the error is
// returned in the `runError` channel which is then closed.
// A caller should listen on `runError` until the `Stop` method
// call fully completes.
//
// If the sequence is already started and not stopped previously,
// the function panics.
func (s *Sequence) Start() (runError <-chan error, startErr error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.internalMutex.Lock()
	defer s.internalMutex.Unlock()

	if s.running {
		panic(fmt.Sprintf("sequence %s already running", s.name))
	}

	var fanInErrorCh <-chan serviceError
	s.fanIn, fanInErrorCh = newErrorsFanIn()

	for _, service := range s.servicesStart {
		serviceString := service.String()

		s.hooks.OnStart(serviceString)
		serviceRunError, err := service.Start()
		s.hooks.OnStarted(serviceString, err)

		if err != nil {
			_ = s.stop()
			return nil, fmt.Errorf("starting %s: %w", serviceString, err)
		}

		s.runningServices[serviceString] = struct{}{}

		s.fanIn.add(serviceString, serviceRunError)
	}

	runErrorCh := make(chan error)
	interceptReady := make(chan struct{})
	s.interceptStop = make(chan struct{})
	s.interceptDone = make(chan struct{})
	go s.interceptRunError(interceptReady, fanInErrorCh, runErrorCh)
	<-interceptReady

	s.running = true

	return runErrorCh, nil
}

// interceptRunError catches an error from the input
// channel, registers the crashed service of the sequence,
// stops other running services and forwards the error
// to the output channel and finally closes this channel.
// If the input error channel is closed, the output channel
// is closed.
func (s *Sequence) interceptRunError(ready chan<- struct{},
	input <-chan serviceError, output chan<- error) {
	defer close(s.interceptDone)
	defer close(output)
	close(ready)

	select {
	case <-s.interceptStop:
	case serviceErr := <-input:
		// Prevent a concurrent entire Start call or Stop call start.
		s.internalMutex.Lock()
		defer s.internalMutex.Unlock()
		delete(s.runningServices, serviceErr.serviceName)
		s.hooks.OnCrash(serviceErr.serviceName, serviceErr.err)
		_ = s.stop()
		output <- &serviceErr
	}
}

// Stop stops running services of the sequence
// in the order specified by the sequence of services.
// If an error occurs for any of the service stop,
// the other running services will still be stopped.
// Only the first non nil service stop error encountered
// is returned, but the hooks can be used to process each
// error returned.
// If the group has already been stopped, the function panics.
func (s *Sequence) Stop() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check the state and stop the intercept goroutine whilst locking
	// the internal mutex to prevent a concurrent modification of the state
	// in `interceptRunError`.
	s.internalMutex.Lock()
	if !s.running {
		panic(fmt.Sprintf("sequence %s already stopped", s.name))
	}
	close(s.interceptStop)

	// unlock to let `interceptRunError` handle an eventual error from
	// the underlying service, and exit.
	// Note Start or another concurrent Stop cannot be called due to the `mutex` lock,
	// so only the terminating `interceptRunError` goroutine can modify the state.
	// There is thus no need to lock the internal mutex below.
	s.internalMutex.Unlock()

	<-s.interceptDone

	if !s.running {
		// The interceptRunError goroutine caught a service run error
		// and stopped all services, whilst this call was waiting on
		// the intercept to be done, so we return nil since the sequence
		// is already stopped.
		return nil
	}

	return s.stop()
}

// stop stops all running services in the sequence.
// If a service fails to stop in the sequence, its error
// is returned but the other services are still stopped.
// Only the first error encountered is returned.
// Hooks can be used to catch each stop error for each service.
func (s *Sequence) stop() (err error) {
	for _, service := range s.servicesStop {
		serviceString := service.String()

		_, running := s.runningServices[serviceString]
		if !running {
			continue
		}

		s.hooks.OnStop(serviceString)
		stopErr := service.Stop()
		s.hooks.OnStopped(serviceString, stopErr)
		if stopErr != nil && err == nil {
			err = fmt.Errorf("stopping %s: %w", serviceString, stopErr)
		}
		delete(s.runningServices, serviceString)
	}

	// Only stop the fan in after stopping all services
	// so it can read and discard any eventual run errors
	// from these whilst we stop them.
	s.fanIn.stop()

	s.running = false

	return err
}
