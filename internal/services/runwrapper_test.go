package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRunWrapper(t *testing.T) {
	t.Parallel()

	run := func(_ context.Context, _ chan<- struct{}, _, _ chan<- error) {}

	wrapper := NewRunWrapper("name", run)

	assert.Equal(t, "name", wrapper.name)
	assert.Equal(t, StateStopped, wrapper.state)
	assert.NotPanics(t, func() {
		wrapper.run(nil, nil, nil, nil)
	})
}

func Test_RunWrapper_String(t *testing.T) {
	t.Parallel()

	wrapper := &RunWrapper{
		name: "name",
	}

	name := wrapper.String()

	assert.Equal(t, "name", name)
}

func Test_RunWrapper_Start(t *testing.T) {
	t.Parallel()

	t.Run("already started", func(t *testing.T) {
		t.Parallel()

		wrapper := &RunWrapper{
			name:  "name",
			state: StateRunning,
		}

		runError, err := wrapper.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, ErrAlreadyStarted)
		assert.EqualError(t, err, "already started")

		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("start error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{},
			runError, stopError chan<- error) {
			runError <- errTest
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("run then stopped", func(t *testing.T) {
		t.Parallel()

		run := func(ctx context.Context, ready chan<- struct{},
			runError, stopError chan<- error) {
			close(ready)
			<-ctx.Done()
			close(stopError)
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()

		require.NoError(t, err)

		assertNoRunError(t, runError)

		err = wrapper.Stop()
		require.NoError(t, err)

		// Check no run error happened during stop
		assertNoRunError(t, runError)

		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("run error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{},
			runError, stopError chan<- error) {
			defer close(stopError)
			close(ready)
			runError <- errTest
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()

		require.NoError(t, err)

		assertRunError(t, runError, errTest)

		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)

		err = wrapper.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed)")
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})
}

func Test_interceptRunError(t *testing.T) {
	t.Parallel()

	t.Run("stop", func(t *testing.T) {
		t.Parallel()

		ready := make(chan struct{})
		stop := make(chan struct{})
		done := make(chan struct{})

		wrapper := &RunWrapper{}
		go wrapper.interceptRunError(ready, stop, done, nil, nil, nil)

		<-ready
		close(stop)
		<-done
	})

	t.Run("closed_runError", func(t *testing.T) {
		t.Parallel()

		ready := make(chan struct{})
		stop := make(chan struct{})
		done := make(chan struct{})
		runErrorIn := make(chan error)
		close(runErrorIn)

		wrapper := &RunWrapper{}
		assert.PanicsWithValue(t, "run error should not be "+
			"closed before writing a single error to it",
			func() {
				wrapper.interceptRunError(ready, stop, done, runErrorIn, nil, nil)
			})

		<-ready
		<-done
	})

	t.Run("error_whilst_stopping", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		ready := make(chan struct{})
		stop := make(chan struct{})
		done := make(chan struct{})
		runErrorIn := make(chan error)
		runErrorOut := make(chan error)
		stopError := make(chan error)

		wrapper := &RunWrapper{
			state: StateStopping,
		}
		go wrapper.interceptRunError(ready, stop, done, runErrorIn, runErrorOut, stopError)

		<-ready
		runErrorIn <- errTest
		err := <-stopError
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed: test error)")
		<-done

		assertNoRunError(t, runErrorOut)
	})

	t.Run("error_caught", func(t *testing.T) {
		t.Parallel()

		ready := make(chan struct{})
		stop := make(chan struct{})
		done := make(chan struct{})
		runErrorIn := make(chan error)
		runErrorOut := make(chan error)

		wrapper := &RunWrapper{
			state: StateRunning,
		}
		go wrapper.interceptRunError(ready, stop, done, runErrorIn, runErrorOut, nil)

		<-ready
		errTest := errors.New("test error")
		runErrorIn <- errTest

		assertRunError(t, runErrorOut, errTest)
		<-done
	})
}

func Test_RunWrapper_Stop(t *testing.T) {
	t.Parallel()

	t.Run("crashed", func(t *testing.T) {
		t.Parallel()

		interceptDone := make(chan struct{})
		close(interceptDone)
		wrapper := &RunWrapper{
			name:          "name",
			state:         StateCrashed,
			interceptDone: interceptDone,
		}

		err := wrapper.Stop()

		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed)")
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("already stopped", func(t *testing.T) {
		t.Parallel()

		wrapper := &RunWrapper{
			name:  "name",
			state: StateStopped,
		}
		err := wrapper.Stop()

		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped")
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("invalid starting state", func(t *testing.T) {
		t.Parallel()

		wrapper := &RunWrapper{
			name:  "name",
			state: StateStarting,
		}

		const expectedPanicMessage = "bad implementation code: " +
			"this code path should be unreachable for the \"starting\" state"
		assert.PanicsWithValue(t, expectedPanicMessage, func() {
			_ = wrapper.Stop()
		})
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("invalid stopping state", func(t *testing.T) {
		t.Parallel()

		wrapper := &RunWrapper{
			name:  "name",
			state: StateStopping,
		}

		const expectedPanicMessage = "bad implementation code: " +
			"this code path should be unreachable for the \"stopping\" state"
		assert.PanicsWithValue(t, expectedPanicMessage, func() {
			_ = wrapper.Stop()
		})
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("stopping with error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		ctx, cancel := context.WithCancel(context.Background())
		stopError := make(chan error)
		go func() { // fake run
			<-ctx.Done()
			stopError <- errTest
		}()

		interceptStop := make(chan struct{})
		interceptDone := make(chan struct{})
		go func() { // fake interceptRunError
			defer close(interceptDone)
			<-interceptStop
		}()

		wrapper := &RunWrapper{
			name:          "name",
			state:         StateRunning,
			interceptStop: interceptStop,
			interceptDone: interceptDone,
			cancel:        cancel,
			stopError:     stopError,
		}

		err := wrapper.Stop()

		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")
		assert.Equal(t, StateStopped, wrapper.state)
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})

	t.Run("stopping without error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		stopError := make(chan error)
		go func() { // fake run
			<-ctx.Done()
			close(stopError)
		}()

		interceptStop := make(chan struct{})
		interceptDone := make(chan struct{})
		go func() { // fake interceptRunError
			defer close(interceptDone)
			<-interceptStop
		}()

		wrapper := &RunWrapper{
			name:          "name",
			state:         StateRunning,
			interceptStop: interceptStop,
			interceptDone: interceptDone,
			cancel:        cancel,
			stopError:     stopError,
		}

		err := wrapper.Stop()

		assert.NoError(t, err)
		assert.Equal(t, StateStopped, wrapper.state)
		assertMutexUnlocked(t, &wrapper.startStopMutex)
		assertMutexUnlocked(t, &wrapper.stateMutex)
	})
}
