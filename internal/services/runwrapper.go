package services

import (
	"context"
	"fmt"
	"sync"
)

// RunFunction is a functional type to simplify a service
// implementation together with `NewRunWrapper`.
//   - `ctx` must be listened on to trigger a stop.
//     Note the `stopError` must be written to when stopping.
//   - `ready` must be closed as soon as the run function
//     has started successfully. Often a simple `close(ready)`
//     at the start of the run body code is enough.
//   - `runError` must have a non-nil error written to it
//     if the run function fails unexpectedly, and the function
//     must promptly `return` right after writing the error.
//     If the function is stopped, the `runError` channel
//     should not be written to nor closed.
//     Alternatively, if a run error is written to this channel,
//     the `stopError` channel should not be written to nor closed.
//     If an error occurs before the run function is ready,
//     the `ready` channel must not be closed.
//     The `runError` channel should be closed after writing an error to it
//     to prevent further writes. Note if an error is written
//     to this channel, the run wrapper service will be considered as
//     crashed.
//   - `stopError` must have an error written to it when the
//     context gets canceled, to signal the stopping result.
//     In the case of no stopping error, the channel must be
//     closed. Otherwise, the error must be written to the
//     channel and the channel should then be closed to prevent
//     further writes. The run function must `return` right after.
//
// A very simple run function template would be:
//
//	func run(ctx context.Context, ready chan<- struct{},
//		runError, stopError chan<- error) {
//		close(ready)
//		select {
//		case <-ctx.Done():
//			// cleanup
//			close(stopError) // successful stop
//			return
//		case err := <-someChannel:
//			if err != nil {
//				runError <- err
//				close(runError)
//				return
//			}
//		}
//	}
type RunFunction func(ctx context.Context,
	ready chan<- struct{}, runError, stopError chan<- error)

// RunWrapper is a service implementation taking care of the many edge cases
// and race conditions that can occur when running a service, and uses
// a user-injected `RunFunction` to run the service.
type RunWrapper struct {
	// Dependencies injected
	name string
	run  RunFunction

	// Internal state
	startStopMutex sync.Mutex
	state          State
	stateMutex     sync.RWMutex

	// Internal fields set at Start
	cancel        context.CancelFunc
	stopError     <-chan error
	interceptStop chan<- struct{}
	interceptDone <-chan struct{}
}

// NewRunWrapper creates a new service wrapper using the
// service name and run function given.
func NewRunWrapper(name string, run RunFunction) *RunWrapper {
	return &RunWrapper{
		name:  name,
		run:   run,
		state: StateStopped,
	}
}

// String returns the name of the service.
func (w *RunWrapper) String() string {
	return w.name
}

// Start starts the service and is thread safe.
// It returns a `runError` channel which the caller should listen
// on to catch an eventual run error from the underlying run function,
// as well as a `startErr` error which can be non-nil if the service
// failed to start.
func (w *RunWrapper) Start() (runError <-chan error, startErr error) {
	// Prevent concurrent Start and Stop calls.
	w.startStopMutex.Lock()
	defer w.startStopMutex.Unlock()

	// Read lock the state in case the service is already running.
	w.stateMutex.RLock()
	state := w.state
	// no need to keep the read lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	w.stateMutex.RUnlock()
	if state == StateRunning {
		return nil, fmt.Errorf("%w", ErrAlreadyStarted)
	}

	w.state = StateStarting

	runErrorToInject := make(chan error)
	runErrorToReturn := make(chan error)
	stopError := make(chan error)
	w.stopError = stopError

	interceptReady := make(chan struct{})
	interceptStop := make(chan struct{})
	w.interceptStop = interceptStop
	interceptDone := make(chan struct{})
	w.interceptDone = interceptDone
	go w.interceptRunError(interceptReady,
		interceptStop, interceptDone,
		runErrorToInject, runErrorToReturn,
		stopError)
	<-interceptReady

	var ctx context.Context
	ctx, w.cancel = context.WithCancel(context.Background())
	runReady := make(chan struct{})
	go w.run(ctx, runReady, runErrorToInject, stopError)

	// Check if there is a run error before the ready channel is closed.
	// That would effectively represent a start error.
	select {
	case <-runReady:
	case startErr = <-runErrorToReturn:
		<-w.interceptDone
		return nil, startErr
	}

	// Lock the state mutex to set the state to running.
	// Any error intercepted from this point would be a run error
	// and no longer a start error, since the run function signaled
	// it was ready by closing the ready channel.
	w.stateMutex.Lock()
	if w.state != StateCrashed {
		// Only set the state to running if the service
		// has not crashed shortly after being ready.
		w.state = StateRunning
	}
	w.stateMutex.Unlock()

	return runErrorToReturn, nil
}

func (w *RunWrapper) interceptRunError(ready chan<- struct{},
	stop <-chan struct{}, done chan<- struct{}, runErrorIn <-chan error,
	runErrorOut, stopError chan<- error) {
	defer close(done)
	close(ready)

	select {
	case <-stop:
		return
	case err, ok := <-runErrorIn:
		if !ok {
			panic("run error should not be closed before writing a single error to it")
		}
		w.stateMutex.Lock()
		if w.state == StateStopping {
			// The run goroutine crashed and the wrapper service is stopping
			// so we send an error to the `stopError` channel and return.
			// The `Stop` method will catch the error from the `stopError`
			// channel and return it as an error from its own call.
			stopError <- fmt.Errorf("%w (crashed: %w)", ErrAlreadyStopped, err)
			close(stopError)
			w.stateMutex.Unlock()
			return
		}
		w.state = StateCrashed
		// unlock mutex since output channel is unbuffered
		w.stateMutex.Unlock()
		runErrorOut <- err
		close(runErrorOut)
	}
}

// Stop stops the service and is thread safe.
// It returns a non-nil error in the following cases:
//   - the underlying run function failed to stop and wrote an error
//     to its `stopError` channel
//   - the service is already stopped
//   - the service is already crashed
func (w *RunWrapper) Stop() (err error) {
	// Prevent concurrent Start and Stop calls.
	w.startStopMutex.Lock()
	defer w.startStopMutex.Unlock()

	w.stateMutex.Lock()
	switch w.state {
	case StateRunning: // continue stopping the service
	case StateCrashed:
		w.stateMutex.Unlock()
		// service is already stopped or stopping from
		// the intercept error goroutine, so just wait for the
		// intercept error goroutine to finish.
		<-w.interceptDone
		// service is now stopped, so return an error indicating
		// it is already stopped.
		return fmt.Errorf("%w (crashed)", ErrAlreadyStopped)
	case StateStopped:
		w.stateMutex.Unlock()
		return fmt.Errorf("%w", ErrAlreadyStopped)
	case StateStarting, StateStopping:
		w.stateMutex.Unlock()
		panic("bad implementation code: " +
			"this code path should be unreachable for the \"" +
			fmt.Sprint(w.state) + "\" state")
	}
	w.state = StateStopping
	w.stateMutex.Unlock()

	w.cancel()
	err = <-w.stopError

	// Stop the intercept error goroutine after the service has been
	// stopped.
	close(w.interceptStop)
	<-w.interceptDone

	w.state = StateStopped

	return err
}
