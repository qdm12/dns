package services

import (
	"fmt"
	"sync"
)

var _ Service = (*Group)(nil)

type Group struct {
	name            string
	services        []Service
	hooks           Hooks
	startStopMutex  *sync.Mutex
	state           State
	stateMutex      *sync.RWMutex
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
		state:           StateStopped,
		stateMutex:      &sync.RWMutex{},
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
// If the group is already running then the function panics.
func (g *Group) Start() (runError <-chan error, startErr error) {
	g.startStopMutex.Lock()
	defer g.startStopMutex.Unlock()

	// Lock the state in case the group is already running.
	g.stateMutex.RLock()
	if g.state == StateRunning {
		panic(fmt.Sprintf("group %s already running", g.name))
	}
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	g.stateMutex.RUnlock()
	g.state = StateStarting

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

	// Hold the state mutex until the intercept run error goroutine is ready
	// and we change the state to running.
	// This is as such because the intercept goroutine may catch a service run error
	// as soon as it starts, and try to set the group state as crashed.
	// With this lock, the goroutine must wait for the mutex unlock below before
	// changing the state to crashed.
	g.stateMutex.Lock()

	runErrorCh := make(chan error)
	interceptReady := make(chan struct{})
	g.interceptStop = make(chan struct{})
	g.interceptDone = make(chan struct{})
	go g.interceptRunError(interceptReady, fanInErrorCh, runErrorCh)
	<-interceptReady

	g.state = StateRunning
	g.stateMutex.Unlock()

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
// If the stop channel triggers, the function returns.
func (g *Group) interceptRunError(ready chan<- struct{},
	input <-chan serviceError, output chan<- error) {
	defer close(g.interceptDone)
	close(ready)

	select {
	case <-g.interceptStop:
	case serviceErr := <-input:
		// Lock the state mutex in case we are stopping
		// or trying to stop the group at the same time.
		g.stateMutex.Lock()
		if g.state == StateStopping {
			// Discard the eventual single service run error
			// fanned-in if we are stopping the group.
			g.stateMutex.Unlock()
			return
		}

		// The first and only service fanned-in run error was
		// caught and we are not currently stopping the group.
		g.state = StateCrashed
		delete(g.runningServices, serviceErr.serviceName)
		g.stateMutex.Unlock()

		g.hooks.OnCrash(serviceErr.serviceName, serviceErr.err)
		_ = g.stop()
		output <- &serviceErr
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

	g.stateMutex.Lock()
	switch g.state {
	case StateRunning: // continue stopping the group
	case StateCrashed:
		g.stateMutex.Unlock()
		// group is already stopped or stopping from
		// the intercept goroutine, so just wait for the
		// intercept goroutine to finish.
		<-g.interceptDone
		return nil
	case StateStopped:
		panic(fmt.Sprintf("bad calling code: group %s already stopped", g.name))
	case StateStarting, StateStopping:
		panic("bad group implementation code: this code path should be unreachable")
	}
	g.state = StateStopping
	g.stateMutex.Unlock()

	err = g.stop()

	// Stop the intercept error goroutine after we stop
	// all the group services. This means the fan in might
	// send one error to the intercept goroutine, but it will
	// discard it since we are in the stopping state.
	// The error fan in takes care of reading and discarding
	// errors from other services once it caught the first error.
	close(g.interceptStop)
	<-g.interceptDone

	g.state = StateStopped

	return err
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

	return err
}
