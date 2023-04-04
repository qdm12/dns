package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_RunWrapper(t *testing.T) {
	t.Parallel()

	t.Run("start_error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			runError <- errTest
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")

		err = wrapper.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed)")
	})

	t.Run("run_then_stopped", func(t *testing.T) {
		t.Parallel()

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			close(ready)
			<-ctx.Done()
			close(stopError) // no error
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.NoError(t, err)

		const waitForErrorDuration = 10 * time.Millisecond
		timer := time.NewTimer(waitForErrorDuration)
		select {
		case <-timer.C:
		case err := <-runError:
			timer.Stop()
			assert.NoError(t, err)
		}

		err = wrapper.Stop()
		assert.NoError(t, err)
	})

	t.Run("run_then_stop_error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			close(ready)
			<-ctx.Done()
			stopError <- errTest
			close(stopError)
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.NoError(t, err)

		const waitForErrorDuration = 10 * time.Millisecond
		timer := time.NewTimer(waitForErrorDuration)
		select {
		case <-timer.C:
		case err := <-runError:
			timer.Stop()
			assert.NoError(t, err)
		}

		err = wrapper.Stop()
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")
	})

	t.Run("fast_run_error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			close(ready)
			runError <- errTest
			close(runError)
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.NoError(t, err)
		assertRunError(t, runError, errTest)

		err = wrapper.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed)")
	})

	t.Run("slow_run_error", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			close(ready)
			timer := time.NewTimer(time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				close(stopError)
			case <-timer.C:
				runError <- errTest
				close(runError)
			}
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.NoError(t, err)

		const waitForRunErrorDuration = time.Second
		timer := time.NewTimer(waitForRunErrorDuration)
		select {
		case err = <-runError:
		case <-timer.C:
			t.Fatalf("no run error received after %s", waitForRunErrorDuration)
		}
		timer.Stop()
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")

		err = wrapper.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed)")
	})

	t.Run("run_error_when_stopping", func(t *testing.T) {
		t.Parallel()
		errTest := errors.New("test error")
		crashTrigger := make(chan struct{})

		run := func(ctx context.Context, ready chan<- struct{}, runError, stopError chan<- error) {
			close(ready)
			<-crashTrigger
			runError <- errTest
			close(runError)
		}

		wrapper := NewRunWrapper("name", run)

		runError, err := wrapper.Start()
		assert.NoError(t, err)
		assertNoRunError(t, runError)

		close(crashTrigger)
		err = wrapper.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "already stopped (crashed: test error)")
	})
}
