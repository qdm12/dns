package services

import (
	"fmt"
	"sync"
)

var _ Service = (*Group)(nil)

type Group struct {
	name            string
	running         bool
	services        []Service
	hooks           Hooks
	startStopMutex  *sync.Mutex
	internalMutex   *sync.Mutex
	fanIn           *errorsFanIn
	runningServices map[string]struct{}
	interceptStop   chan struct{}
	interceptDone   chan struct{}
}

func NewGroup(settings GroupSettings) (group *Group, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	services := make([]Service, len(settings.Services))
	copy(services, settings.Services)

	return &Group{
		name:            settings.Name,
		services:        services,
		hooks:           settings.Hooks,
		startStopMutex:  &sync.Mutex{},
		internalMutex:   &sync.Mutex{},
		runningServices: make(map[string]struct{}),
	}, nil
}

func (g *Group) String() string {
	if g.name == "" {
		return "group"
	}
	return "group " + g.name
}

// Start starts services specified in parallel.
//
// If a service fails to start, the `startErr` is returned
// and all other running services are stopped.
//
// If a service fails after `Start` returns without error,
// all other running services are stopped and the error is
// sent in the `runError` channel which is then closed.
// A caller should listen on `runError` until the `Stop` method
// call fully completes, since a run error can theoretically happen
// at the same time the caller calls `Stop` on the group.
//
// If the group of services is already started and not
// stopped previously, the function panics.
func (g *Group) Start() (runError <-chan error, startErr error) {
	g.startStopMutex.Lock()
	defer g.startStopMutex.Unlock()

	g.internalMutex.Lock()
	defer g.internalMutex.Unlock()

	if g.running {
		panic(fmt.Sprintf("group %s already running", g.name))
	}

	var fanInErrorCh <-chan serviceError
	g.fanIn, fanInErrorCh = newErrorsFanIn()

	runErrorChannels := make(map[string]<-chan error, len(g.services))
	startErrorCh := make(chan *serviceError)
	runErrorMapMutex := new(sync.Mutex)
	for _, service := range g.services {
		serviceString := service.String()
		go startGroupedServiceAsync(service, serviceString, g.hooks,
			startErrorCh, runErrorChannels, runErrorMapMutex)
		// assume all the services are going to be running
		g.runningServices[serviceString] = struct{}{}
	}

	// Collect eventual start error and wait for all services
	// to be started or failed to start.
	for range g.services {
		serviceErr := <-startErrorCh
		if serviceErr == nil {
			continue
		}

		delete(g.runningServices, serviceErr.serviceName)

		if startErr == nil {
			startErr = serviceErr
		}
	}

	if startErr != nil {
		_ = g.stop()
		return nil, startErr
	}

	for serviceString, runError := range runErrorChannels {
		g.fanIn.add(serviceString, runError)
	}

	runErrorCh := make(chan error)
	interceptReady := make(chan struct{})
	g.interceptStop = make(chan struct{})
	g.interceptDone = make(chan struct{})
	go g.interceptRunError(interceptReady, fanInErrorCh, runErrorCh)
	<-interceptReady

	g.running = true

	return runErrorCh, nil
}

func startGroupedServiceAsync(service Starter, serviceString string,
	hooks Hooks, startErrorCh chan<- *serviceError,
	runErrorChannels map[string]<-chan error, mutex *sync.Mutex) {
	hooks.OnStart(serviceString)
	runError, err := service.Start()
	hooks.OnStarted(serviceString, err)

	if err != nil {
		startErrorCh <- &serviceError{
			format:      errorFormatStart,
			serviceName: serviceString,
			err:         err,
		}
		return
	}

	mutex.Lock()
	runErrorChannels[serviceString] = runError
	mutex.Unlock()
	startErrorCh <- nil
}

// interceptRunError, if it catches an error from the input
// channel, registers the crashed service of the group,
// stops other running services and forwards the error
// to the output channel and finally closes this channel.
// If the stop channel triggers, the function returns
// and closes the output channel.
func (g *Group) interceptRunError(ready chan<- struct{},
	input <-chan serviceError, output chan<- error) {
	defer close(g.interceptDone)
	close(ready)

	select {
	case <-g.interceptStop:
	case serviceErr := <-input:
		// Prevent a concurrent entire Start call or Stop call start.
		g.internalMutex.Lock()
		defer g.internalMutex.Unlock()
		delete(g.runningServices, serviceErr.serviceName)
		_ = g.stop()
		output <- serviceErr
		close(output)
	}
}

// Stop stops running services of the group in parallel.
// If an error occurs for any of the service stop,
// the other running services will still be stopped.
// Only the first non nil service stop error encountered
// is returned, but the hooks can be used to process each
// error returned.
// If the group has already been stopped, the function panics.
func (g *Group) Stop() (err error) {
	g.startStopMutex.Lock()
	defer g.startStopMutex.Unlock()

	// Check the state and stop the intercept goroutine whilst locking
	// the internal mutex to prevent a concurrent modification of the state
	// in `interceptRunError`.
	g.internalMutex.Lock()

	if !g.running {
		panic(fmt.Sprintf("group %s already stopped", g.name))
	}
	close(g.interceptStop)

	// unlock to let `interceptRunError` handle an eventual error from
	// the underlying service, and exit.
	// Note Start or another concurrent Stop cannot be called due to the `mutex` lock,
	// so only the terminating `interceptRunError` goroutine can modify the state.
	// There is thus no need to lock the internal mutex below.
	g.internalMutex.Unlock()

	<-g.interceptDone

	if !g.running {
		// The interceptRunError goroutine caught a service run error
		// and stopped all services, whilst this call was waiting on
		// the intercept to be done, so we return nil since the group
		// is already stopped.
		return nil
	}

	return g.stop()
}

// stop stops all running services in the group of services.
// If a service fails to stop in the group, its error
// is returned but the other services are still stopped.
// Only the first error encountered is returned.
// Hooks can be used to catch each stop error for each service.
func (g *Group) stop() (err error) {
	stopErrors := make(chan serviceError)
	var runningCount uint

	for _, service := range g.services {
		serviceString := service.String()

		_, running := g.runningServices[serviceString]
		if !running {
			continue
		}
		runningCount++

		go func(service Stopper, serviceString string, stopErrors chan<- serviceError) {
			g.hooks.OnStop(serviceString)
			err := service.Stop()
			g.hooks.OnStopped(serviceString, err)
			stopErrors <- serviceError{
				format:      errorFormatStop,
				serviceName: serviceString,
				err:         err,
			}
		}(service, serviceString, stopErrors)
	}

	for i := uint(0); i < runningCount; i++ {
		stopErr := <-stopErrors
		if stopErr.err != nil && err == nil {
			err = stopErr
		}

		delete(g.runningServices, stopErr.serviceName)
	}

	// Only stop the fan in after stopping all services
	// so it can read and discard any eventual run errors
	// from these whilst we stop them.
	g.fanIn.stop()

	g.running = false

	return err
}
