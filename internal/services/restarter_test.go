package services

import (
	"errors"
	"sync"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/internal/services/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRestarter(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	dummyService := NewMockService(ctrl)
	dummyService.EXPECT().String().Return("A").AnyTimes()

	testCases := map[string]struct {
		settings    RestarterSettings
		restarter   *Restarter
		errSentinel error
		errMessage  string
	}{
		"missing service": {
			errSentinel: ErrNoService,
			errMessage:  "validating settings: no service specified",
		},
		"minimal settings": {
			settings: RestarterSettings{
				Service: dummyService,
			},
			restarter: &Restarter{
				service:        dummyService,
				hooks:          hooks.NewNoop(),
				startStopMutex: &sync.Mutex{},
				startedMutex:   &sync.Mutex{},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			restarter, err := NewRestarter(testCase.settings)

			assert.ErrorIs(t, err, testCase.errSentinel)
			if testCase.errSentinel != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.restarter, restarter)
		})
	}
}

func Test_Restarter_String(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	service := NewMockService(ctrl)
	service.EXPECT().String().Return("A")

	restarter := &Restarter{
		service: service,
	}

	s := restarter.String()

	assert.Equal(t, "A", s)
}

func Test_Restarter_Start(t *testing.T) {
	t.Parallel()

	t.Run("panic if already running", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		service := NewMockService(ctrl)
		service.EXPECT().String().Return("A")

		restarter := &Restarter{
			service:        service,
			running:        true,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
		}

		assert.PanicsWithValue(t,
			"restarter for A already running",
			func() {
				_, _ = restarter.Start()
			})
	})

	t.Run("service first start error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		service := NewMockService(ctrl)
		service.EXPECT().String().Return("A") // Start method
		hooks.EXPECT().OnStart("A")
		errTest := errors.New("test error")
		service.EXPECT().Start().Return(nil, errTest)
		hooks.EXPECT().OnStarted("A", errTest)

		settings := RestarterSettings{
			Service: service,
			Hooks:   hooks,
		}

		restarter, err := NewRestarter(settings)
		require.NoError(t, err)

		runError, err := restarter.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")
	})

	t.Run("restart service multiple times", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)
		service := NewMockService(ctrl)

		service.EXPECT().String().Return("A") // Start method
		hooks.EXPECT().OnStart("A")
		runErrorService := make(chan error, 1)
		service.EXPECT().Start().Return(runErrorService, nil)
		hooks.EXPECT().OnStarted("A", nil)

		settings := RestarterSettings{
			Service: service,
			Hooks:   hooks,
		}

		restarter, err := NewRestarter(settings)
		require.NoError(t, err)

		runError, err := restarter.Start()
		require.NoError(t, err)
		require.True(t, restarter.running)

		const numberOfRestarts = 5
		wg := new(sync.WaitGroup)
		wg.Add(numberOfRestarts)
		for i := 0; i < numberOfRestarts; i++ {
			// Restart expectations
			errTest := errors.New("test error")
			hooks.EXPECT().OnCrash("A", errTest)
			hooks.EXPECT().OnStart("A")
			nextRunErrorService := make(chan error, 1)
			service.EXPECT().Start().Return(nextRunErrorService, nil)
			hooks.EXPECT().OnStarted("A", nil).Do(func(_ string, _ error) {
				wg.Done()
			})

			// Trigger restart
			runErrorService <- errTest
			close(runErrorService)

			// No error should be sent in the restart run error channel.
			select {
			case err := <-runError:
				require.NoError(t, err)
			default:
			}

			require.True(t, restarter.running)

			runErrorService = nextRunErrorService
		}

		// Wait for all restarts to complete before calling Stop or some
		// of the expectations might not be met.
		wg.Wait()

		service.EXPECT().String().Return("A") // Stop method
		hooks.EXPECT().OnStop("A")
		service.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("A", nil)
		err = restarter.Stop()
		require.NoError(t, err)
		assert.False(t, restarter.running)
	})

	t.Run("restart service fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)
		service := NewMockService(ctrl)

		service.EXPECT().String().Return("A") // Start method
		hooks.EXPECT().OnStart("A")
		runErrorService := make(chan error, 1)
		service.EXPECT().Start().Return(runErrorService, nil)
		hooks.EXPECT().OnStarted("A", nil)

		settings := RestarterSettings{
			Service: service,
			Hooks:   hooks,
		}

		restarter, err := NewRestarter(settings)
		require.NoError(t, err)

		runError, err := restarter.Start()
		require.NoError(t, err)
		require.True(t, restarter.running)

		// Restart expectations
		errTest := errors.New("test error")
		hooks.EXPECT().OnCrash("A", errTest)
		hooks.EXPECT().OnStart("A")
		errStartTest := errors.New("test error")
		service.EXPECT().Start().Return(nil, errStartTest)
		hooks.EXPECT().OnStarted("A", errStartTest)

		// Trigger restart
		runErrorService <- errTest
		close(runErrorService)

		err = <-runError
		assert.ErrorIs(t, err, errStartTest)
		assert.EqualError(t, err, "restarting after crash: test error")

		<-runError
		assert.False(t, restarter.running)
	})
}

func Test_Restarter_interceptRunError(t *testing.T) {
	t.Parallel()

	t.Run("stop", func(t *testing.T) {
		t.Parallel()

		restarter := Restarter{
			interceptStop: make(chan struct{}),
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		output := make(chan error)
		go restarter.interceptRunError(ready, "", nil, output)
		<-ready
		close(restarter.interceptStop)
		<-restarter.interceptDone
	})

	t.Run("restart success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		service := NewMockService(ctrl)
		hooks := NewMockHooks(ctrl)

		restarter := Restarter{
			service:        service,
			hooks:          hooks,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
			interceptStop:  make(chan struct{}),
			interceptDone:  make(chan struct{}),
		}
		ready := make(chan struct{})
		const serviceName = "A"
		input := make(chan error)
		output := make(chan error)
		go restarter.interceptRunError(ready,
			serviceName, input, output)
		<-ready

		errTest := errors.New("test error")
		hooks.EXPECT().OnCrash(serviceName, errTest)
		hooks.EXPECT().OnStart(serviceName)
		service.EXPECT().Start().Return(nil, nil)
		hooks.EXPECT().OnStarted(serviceName, nil)
		input <- errTest
		close(input)

		close(restarter.interceptStop)
		<-restarter.interceptDone
	})

	t.Run("restart failure", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		service := NewMockService(ctrl)
		hooks := NewMockHooks(ctrl)

		restarter := Restarter{
			service:        service,
			hooks:          hooks,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
			interceptStop:  make(chan struct{}),
			interceptDone:  make(chan struct{}),
		}
		ready := make(chan struct{})
		const serviceName = "A"
		input := make(chan error)
		output := make(chan error)
		go restarter.interceptRunError(ready,
			serviceName, input, output)
		<-ready

		errTest := errors.New("test error")
		hooks.EXPECT().OnCrash(serviceName, errTest)
		hooks.EXPECT().OnStart(serviceName)
		errStartTest := errors.New("test start error")
		service.EXPECT().Start().Return(nil, errStartTest)
		hooks.EXPECT().OnStarted(serviceName, errStartTest)
		input <- errTest
		close(input)
		err := <-output
		assert.ErrorIs(t, err, errStartTest)
		assert.EqualError(t, err, "restarting after crash: test start error")
		<-output
		<-restarter.interceptDone
	})
}

func Test_Restarter_Stop(t *testing.T) {
	t.Parallel()

	t.Run("already stopped", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		service := NewMockService(ctrl)
		service.EXPECT().String().Return("A")

		restarter := Restarter{
			service:        service,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
		}
		assert.PanicsWithValue(t, "restarter for A already stopped", func() {
			_ = restarter.Stop()
		})
	})

	t.Run("started", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		service := NewMockService(ctrl)
		service.EXPECT().String().Return("A")
		hooks.EXPECT().OnStop("A")
		errTest := errors.New("test error")
		service.EXPECT().Stop().Return(errTest)
		hooks.EXPECT().OnStopped("A", errTest)

		restarter := Restarter{
			running:        true,
			service:        service,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
			hooks:          hooks,
			interceptStop:  make(chan struct{}),
			interceptDone:  make(chan struct{}),
		}

		// Simulate interceptRunError exiting from stop signal.
		go func() {
			<-restarter.interceptStop
			close(restarter.interceptDone)
		}()

		err := restarter.Stop()
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "test error")
	})

	t.Run("crash during stop", func(t *testing.T) {
		t.Parallel()

		restarter := Restarter{
			running:        true,
			startStopMutex: &sync.Mutex{},
			startedMutex:   &sync.Mutex{},
			interceptStop:  make(chan struct{}),
			interceptDone:  make(chan struct{}),
		}

		// Simulate interceptRunError handling a service crash
		go func() {
			<-restarter.interceptStop
			restarter.running = false
			close(restarter.interceptDone)
		}()

		err := restarter.Stop()
		assert.NoError(t, err)
	})
}
