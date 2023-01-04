package services

import (
	"fmt"
	"sync"
)

var _ Service = (*Restarter)(nil)

type Restarter struct {
	service        Service
	hooks          Hooks
	startStopMutex sync.Mutex
	state          State
	stateMutex     sync.RWMutex
	interceptStop  chan struct{}
	interceptDone  chan struct{}
}

func NewRestarter(settings RestarterSettings) (restarter *Restarter, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Restarter{
		service: settings.Service,
		hooks:   settings.Hooks,
		state:   StateStopped,
	}, nil
}

func (r *Restarter) String() string {
	return r.service.String()
}

// Start starts the underlying service.
//
// If the underlying service fails to start, the `startErr` is returned.
//
// If the underlying service fails after this method call returns
// without error, it is automatically restarted and no error is emitted
// in the `runError` channel.
//
// If a subsequent service start fails, the start error is sent in the
// `runError` channel, this channel is closed and the restarter stops.
// A caller should listen on `runError` until the `Stop` method
// call fully completes, since a run error can theoretically happen
// at the same time the caller calls `Stop` on the restarter.
//
// If the restarter is already started and not stopped previously,
// the function panics.
func (r *Restarter) Start() (runError <-chan error, startErr error) {
	// Prevent concurrent Stop and Start calls.
	r.startStopMutex.Lock()
	defer r.startStopMutex.Unlock()

	// Lock the state in case the restarter is already running.
	r.stateMutex.RLock()
	if r.state == StateRunning {
		panic(fmt.Sprintf("restarter for %s already running", r.service))
	}
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	r.stateMutex.RUnlock()
	r.state = StateStarting

	serviceString := r.service.String()

	r.hooks.OnStart(serviceString)
	serviceRunError, startErr := r.service.Start()
	r.hooks.OnStarted(serviceString, startErr)

	if startErr != nil {
		return nil, startErr
	}

	// Hold the state mutex until the intercept run error goroutine is ready
	// and we change the state to running.
	// This is as such because the intercept goroutine may catch a service run error
	// as soon as it starts, and try to set the restarter state as crashed.
	// With this lock, the goroutine must wait for the mutex unlock below before
	// changing the state to crashed.
	r.stateMutex.Lock()

	interceptReady := make(chan struct{})
	runErrorCh := make(chan error)
	r.interceptStop = make(chan struct{})
	r.interceptDone = make(chan struct{})
	go r.interceptRunError(interceptReady, serviceString,
		serviceRunError, runErrorCh)
	<-interceptReady

	r.state = StateRunning
	r.stateMutex.Unlock()

	return runErrorCh, nil
}

func (r *Restarter) interceptRunError(ready chan<- struct{},
	serviceName string, input <-chan error, output chan<- error) {
	defer close(r.interceptDone)
	close(ready)

	for {
		select {
		case <-r.interceptStop:
			return
		case err := <-input:
			// Lock the state mutex in case we are stopping
			// or trying to stop the restarter at the same time.
			r.stateMutex.Lock()
			if r.state == StateStopping {
				// Discard the eventual single service run error
				// if we are stopping the restarter.
				r.stateMutex.Unlock()
				return
			}

			r.hooks.OnCrash(serviceName, err)

			r.hooks.OnStart(serviceName)
			var startErr error
			input, startErr = r.service.Start()
			r.hooks.OnStarted(serviceName, startErr)

			if startErr != nil {
				r.state = StateCrashed
				r.stateMutex.Unlock()
				output <- fmt.Errorf("restarting after crash: %w", startErr)
				close(output)
				return
			}
			r.state = StateRunning
			r.stateMutex.Unlock()
		}
	}
}

// Stop stops the underlying service and the internal
// run error restart-watcher goroutine.
// If the restarter has already been stopped, the function panics.
func (r *Restarter) Stop() (err error) {
	r.startStopMutex.Lock()
	defer r.startStopMutex.Unlock()

	r.stateMutex.Lock()
	switch r.state {
	case StateRunning: // continue stopping the restarter
	case StateCrashed:
		// service crashed and failed to restart, just wait
		// for the intercept goroutine to finish.
		<-r.interceptDone
		return nil
	case StateStopped:
		panic(fmt.Sprintf("bad calling code: restarter for %s already stopped", r.service))
	case StateStarting, StateStopping:
		panic("bad restarter implementation code: this code path should be unreachable")
	}
	r.state = StateStopping
	r.stateMutex.Unlock()

	serviceString := r.service.String()

	r.hooks.OnStop(serviceString)
	err = r.service.Stop()
	r.hooks.OnStopped(serviceString, err)

	// Stop the intercept error goroutine after we stop
	// the restarter underlying service.
	close(r.interceptStop)
	<-r.interceptDone

	r.state = StateStopped

	return err
}
