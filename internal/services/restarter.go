package services

import (
	"fmt"
	"sync"
)

var _ Service = (*Restarter)(nil)

type Restarter struct {
	running       bool
	service       Service
	hooks         Hooks
	mutex         *sync.Mutex
	startedMutex  *sync.Mutex
	interceptStop chan struct{}
	interceptDone chan struct{}
}

func NewRestarter(settings RestarterSettings) (restarter *Restarter, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Restarter{
		service:      settings.Service,
		hooks:        settings.Hooks,
		mutex:        &sync.Mutex{},
		startedMutex: &sync.Mutex{},
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
// `runError` channel and the restarter stops.
//
// The `runError` channel is closed when the restarter is stopped.
//
// If the restarter is already started and not stopped previously,
// the function panics.
func (r *Restarter) Start() (runError <-chan error, startErr error) {
	// Prevent concurrent Stop and Start calls.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.startedMutex.Lock()
	defer r.startedMutex.Unlock()

	if r.running {
		panic(fmt.Sprintf("restarter for %s already running", r.service))
	}

	serviceString := r.service.String()

	r.hooks.OnStart(serviceString)
	serviceRunError, startErr := r.service.Start()
	r.hooks.OnStarted(serviceString, startErr)

	if startErr != nil {
		return nil, startErr
	}

	interceptReady := make(chan struct{})
	runErrorCh := make(chan error)
	r.interceptStop = make(chan struct{})
	r.interceptDone = make(chan struct{})
	go r.interceptRunError(interceptReady, serviceString,
		serviceRunError, runErrorCh)
	<-interceptReady

	r.running = true

	return runErrorCh, nil
}

func (r *Restarter) interceptRunError(ready chan<- struct{},
	serviceName string, input <-chan error, output chan<- error) {
	defer close(r.interceptDone)
	defer close(output)
	close(ready)

	for {
		select {
		case <-r.interceptStop:
			return
		case err := <-input:
			// Prevent a concurrent entire Start call or Stop call start.
			r.startedMutex.Lock()

			r.hooks.OnCrash(serviceName, err)

			r.hooks.OnStart(serviceName)
			var startErr error
			input, startErr = r.service.Start()
			r.hooks.OnStarted(serviceName, startErr)

			if startErr != nil {
				r.running = false
				output <- fmt.Errorf("restarting after crash: %w", startErr)
				r.startedMutex.Unlock()
				return
			}
			r.startedMutex.Unlock()
		}
	}
}

// Stop stops the underlying service and the internal
// run error restart-watcher goroutine.
// It closes the `runError` channel returned by `Start`.
// If the group has already been stopped, the function panics.
func (r *Restarter) Stop() (err error) {
	// Prevent concurrent Stop and Start calls.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check the state and stop the intercept goroutine whilst locking
	// the started mutex to prevent a concurrent modification of the state
	// in `interceptRunError`.
	r.startedMutex.Lock()
	if !r.running {
		panic(fmt.Sprintf("restarter for %s already stopped", r.service))
	}
	close(r.interceptStop)

	// unlock to let `interceptRunError` handle an eventual error from
	// the underlying service, and exit.
	// Note Start or another concurrent Stop cannot be called due to the `mutex` lock,
	// so only the terminating `interceptRunError` goroutine can modify the state.
	// There is thus no need to lock the started mutex below.
	r.startedMutex.Unlock()

	<-r.interceptDone

	if !r.running {
		// The interceptRunError goroutine had failed restarting the
		// underlying service and set the state to stopped, whilst this
		// call was waiting on the intercept to be done, so we
		// return nil since the restarter is already stopped.
		return nil
	}

	serviceString := r.service.String()

	r.hooks.OnStop(serviceString)
	err = r.service.Stop()
	r.hooks.OnStopped(serviceString, err)

	r.running = false
	return err
}
