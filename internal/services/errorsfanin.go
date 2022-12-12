package services

import (
	"sync"
)

// errorsFanIn takes care of fanning run errors from
// different error channels to a single error channel.
// It writes only the first run error read to the output
// channel returned by its constructor, and discards other
// errors received until the fan in is stopped.
// Each service run error channel should send one error at most.
type errorsFanIn struct {
	runErrors          []<-chan error
	serviceToFaninStop []chan<- struct{}
	serviceToFaninDone []<-chan struct{}
	output             chan serviceError
	runErrorMutex      *sync.Mutex
}

// newErrorsFanIn returns a new errors fan in object
// together with the output channel for the first service error.
func newErrorsFanIn() (fanIn *errorsFanIn, reader <-chan serviceError) {
	output := make(chan serviceError)
	return &errorsFanIn{
		output:        output,
		runErrorMutex: &sync.Mutex{},
	}, output
}

// add adds a run error receiving channel to the fan in mechanism
// for the particular service string given.
// This is NOT thread safe to call.
// Only the first error received is read from the given run
// error channel, other errors are not listened for.
func (e *errorsFanIn) add(service string, runError <-chan error) {
	e.runErrors = append(e.runErrors, runError)
	stopCh := make(chan struct{})
	e.serviceToFaninStop = append(e.serviceToFaninStop, stopCh)
	fanInDone := make(chan struct{})
	e.serviceToFaninDone = append(e.serviceToFaninDone, fanInDone)
	ready := make(chan struct{})

	go e.fanIn(service, runError, ready, stopCh, fanInDone)
	<-ready
}

func (e *errorsFanIn) fanIn(service string, input <-chan error,
	ready chan<- struct{}, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	close(ready)

	select {
	case <-stop:
		// Drain input so the service doesn't hang if it crashed
		// at the same time as we're stopping the fan in.
		select {
		case <-input:
		default:
		}
	case err, ok := <-input:
		if !ok {
			// The service channel cannot be closed before we read
			// at least one error from it.
			panic("run error service channel closed unexpectedly")
		}

		// Check we are not racing with the stop signal.
		select {
		case <-stop:
			return
		default:
		}

		// Use a mutex to prevent concurrent processing of run errors.
		e.runErrorMutex.Lock()
		defer e.runErrorMutex.Unlock()

		if isOutputClosed(e.output) {
			// if the output is closed, we already received a run error
			// previously so discard this error and do not write to the
			// closed output channel.
			// This is especially useful to drain the possibly unbuffered
			// input channel so the service writing to it is not stuck waiting
			// for a reader to read its error.
			return
		}

		serviceErr := serviceError{
			format:      errorFormatCrash,
			serviceName: service,
			err:         err,
		}
		e.output <- serviceErr
		close(e.output)
	}
}

func isOutputClosed(output <-chan serviceError) (closed bool) {
	select {
	case _, ok := <-output:
		return !ok
	default:
		return false
	}
}

// stop stops the fan in and closes the reader channel
// returned by the fan in constructor, if it has not been closed
// already.
// Note this should be called only when all services have been stopped
// so they don't get stuck trying to write to their run error channel
// with no channel reader anymore.
func (e *errorsFanIn) stop() {
	for i := 0; i < len(e.runErrors); i++ {
		close(e.serviceToFaninStop[i])
		<-e.serviceToFaninDone[i]
	}
	if !isOutputClosed(e.output) {
		close(e.output)
	}
}
