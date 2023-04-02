package services

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/internal/services/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewGroup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	dummyService := NewMockService(ctrl)
	dummyService.EXPECT().String().Return("A").AnyTimes()

	testCases := map[string]struct {
		settings    GroupSettings
		group       *Group
		errSentinel error
		errMessage  string
	}{
		"empty settings": {
			errSentinel: ErrNoService,
			errMessage:  "validating settings: no service specified",
		},
		"full settings": {
			settings: GroupSettings{
				Name:     "name",
				Services: []Service{dummyService},
				Hooks:    hooks.NewWithLog(nil),
			},
			group: &Group{
				name:            "name",
				services:        []Service{dummyService},
				hooks:           hooks.NewWithLog(nil),
				runningServices: map[string]struct{}{},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			group, err := NewGroup(testCase.settings)

			assert.ErrorIs(t, err, testCase.errSentinel)
			if testCase.errSentinel != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.group, group)
		})
	}
}

func Test_Group_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		group    *Group
		expected string
	}{
		"empty name": {
			group:    &Group{},
			expected: "group",
		},
		"set name": {
			group: &Group{
				name: "A",
			},
			expected: "group A",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := testCase.group.String()

			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func Test_Group_Start(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	t.Run("error if already running", func(t *testing.T) {
		t.Parallel()

		group := &Group{
			name:  "name",
			state: StateRunning,
		}

		_, err := group.Start()

		assert.ErrorIs(t, err, ErrAlreadyStarted)
		assert.EqualError(t, err, "group name: already started")
	})

	t.Run("one service of two start error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A").Times(2) // settings validation
		hooks.EXPECT().OnStart("A")
		serviceA.EXPECT().String().Return("A") // Start method
		serviceA.EXPECT().Start().Return(nil, errTest)
		hooks.EXPECT().OnStarted("A", errTest)
		serviceA.EXPECT().String().Return("A") // stop method

		serviceB := NewMockService(ctrl)
		serviceB.EXPECT().String().Return("B").Times(2) // settings validation
		hooks.EXPECT().OnStart("B")
		serviceB.EXPECT().String().Return("B") // Start method
		serviceB.EXPECT().Start().Return(nil, nil)
		hooks.EXPECT().OnStarted("B", nil)
		serviceB.EXPECT().String().Return("B") // stop method
		hooks.EXPECT().OnStop("B")
		serviceB.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("B", nil)

		settings := GroupSettings{
			Services: []Service{serviceA, serviceB},
			Hooks:    hooks,
		}

		group, err := NewGroup(settings)
		require.NoError(t, err)

		runError, err := group.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "starting A: test error")
	})

	t.Run("two services of two start error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A").Times(2) // settings validation
		hooks.EXPECT().OnStart("A")
		serviceA.EXPECT().String().Return("A") // Start method
		serviceA.EXPECT().Start().Return(nil, errTest)
		hooks.EXPECT().OnStarted("A", errTest)
		serviceA.EXPECT().String().Return("A") // stop method

		serviceB := NewMockService(ctrl)
		serviceB.EXPECT().String().Return("B").Times(2) // settings validation
		hooks.EXPECT().OnStart("B")
		serviceB.EXPECT().String().Return("B") // Start method
		serviceB.EXPECT().Start().Return(nil, errTest)
		hooks.EXPECT().OnStarted("B", errTest)
		serviceB.EXPECT().String().Return("B") // stop method

		settings := GroupSettings{
			Services: []Service{serviceA, serviceB},
			Hooks:    hooks,
		}

		group, err := NewGroup(settings)
		require.NoError(t, err)

		runError, err := group.Start()

		assert.Nil(t, runError)
		require.ErrorIs(t, err, errTest)
		assert.Regexp(t, "starting (A|B): test error", err.Error())
	})

	t.Run("start success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A").Times(2) // settings validation
		hooks.EXPECT().OnStart("A")
		serviceA.EXPECT().String().Return("A") // Start method
		runErrorA := make(chan error)
		serviceA.EXPECT().Start().Return(runErrorA, nil)
		hooks.EXPECT().OnStarted("A", nil)

		serviceB := NewMockService(ctrl)
		serviceB.EXPECT().String().Return("B").Times(2) // settings validation
		hooks.EXPECT().OnStart("B")
		serviceB.EXPECT().String().Return("B") // Start method
		runErrorB := make(chan error)
		serviceB.EXPECT().Start().Return(runErrorB, nil)
		hooks.EXPECT().OnStarted("B", nil)

		settings := GroupSettings{
			Services: []Service{serviceA, serviceB},
			Hooks:    hooks,
		}

		group, err := NewGroup(settings)
		require.NoError(t, err)

		runError, err := group.Start()

		require.NoError(t, err)
		require.NotNil(t, runError)

		select {
		case err := <-runError:
			assert.NoError(t, err)
		default:
		}

		// Expectations for the group stop call.
		serviceA.EXPECT().String().Return("A") // stop method
		serviceB.EXPECT().String().Return("B") // stop method
		hooks.EXPECT().OnStop("B")
		serviceB.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("B", nil)
		hooks.EXPECT().OnStop("A")
		serviceA.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("A", nil)

		err = group.Stop()
		assert.NoError(t, err)
	})

	t.Run("run error after start", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A").Times(4)
		hooks.EXPECT().OnStart("A")
		runErrorA := make(chan error)
		serviceA.EXPECT().Start().Return(runErrorA, nil)
		hooks.EXPECT().OnStarted("A", nil)

		serviceB := NewMockService(ctrl)
		serviceB.EXPECT().String().Return("B").Times(4)
		hooks.EXPECT().OnStart("B")
		runErrorB := make(chan error)
		serviceB.EXPECT().Start().Return(runErrorB, nil)
		hooks.EXPECT().OnStarted("B", nil)

		settings := GroupSettings{
			Services: []Service{serviceA, serviceB},
			Hooks:    hooks,
		}

		group, err := NewGroup(settings)
		require.NoError(t, err)

		runError, startErr := group.Start()
		require.NoError(t, startErr)

		// Stop service B since A crashes
		hooks.EXPECT().OnStop("B")
		serviceB.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("B", nil)

		hooks.EXPECT().OnCrash("A", errTest)
		runErrorA <- errTest
		err = <-runError
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "A crashed: test error")

		_, ok := <-runError
		assert.False(t, ok)
	})
}

func Test_Group_interceptRunError(t *testing.T) {
	t.Parallel()

	t.Run("stop", func(t *testing.T) {
		t.Parallel()

		group := Group{
			interceptStop: make(chan struct{}),
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		output := make(chan error)
		go group.interceptRunError(ready, nil, output)
		<-ready

		close(group.interceptStop)
		<-group.interceptDone
	})

	t.Run("stopping in progress", func(t *testing.T) {
		t.Parallel()

		group := Group{
			state:         StateStopping,
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		input := make(chan serviceError)
		output := make(chan error)
		go group.interceptRunError(ready, input, output)
		<-ready

		input <- serviceError{}

		<-group.interceptDone
	})

	t.Run("one of two services crash", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		fanIn, _ := newErrorsFanIn()
		hooks := NewMockHooks(ctrl)
		serviceA := NewMockService(ctrl)
		serviceB := NewMockService(ctrl)

		errTest := errors.New("test error")
		hooks.EXPECT().OnCrash("A", errTest)

		// Expectations for stop method call.
		serviceA.EXPECT().String().Return("A")
		serviceB.EXPECT().String().Return("B")
		hooks.EXPECT().OnStop("B")
		errStop := errors.New("stop error")
		serviceB.EXPECT().Stop().Return(errStop) // ignored error
		hooks.EXPECT().OnStopped("B", errStop)

		group := &Group{
			runningServices: map[string]struct{}{
				"A": {},
				"B": {},
			},
			services:      []Service{serviceA, serviceB},
			fanIn:         fanIn,
			hooks:         hooks,
			state:         StateRunning,
			interceptStop: make(chan struct{}),
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		input := make(chan serviceError)
		output := make(chan error)

		go group.interceptRunError(ready, input, output)

		<-ready

		input <- serviceError{
			format:      errorFormatCrash,
			serviceName: "A",
			err:         errTest,
		}

		err := <-output
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "A crashed: test error")

		_, ok := <-output
		assert.False(t, ok)

		<-group.interceptDone

		expectedGroup := &Group{
			runningServices: map[string]struct{}{},
			services:        []Service{serviceA, serviceB},
			fanIn:           fanIn,
			hooks:           hooks,
			state:           StateCrashed,
		}
		group.interceptStop = nil
		group.interceptDone = nil
		assert.Equal(t, expectedGroup, group)
	})
}

func Test_Group_Stop(t *testing.T) {
	t.Parallel()

	t.Run("already stopped returns an error", func(t *testing.T) {
		t.Parallel()

		group := Group{
			name:  "name",
			state: StateStopped,
		}
		err := group.Stop()
		assert.ErrorIs(t, err, ErrAlreadyStopped)
		assert.EqualError(t, err, "group name: already stopped")
	})

	t.Run("in starting state", func(t *testing.T) {
		t.Parallel()

		group := Group{
			name:  "name",
			state: StateStarting,
		}
		assert.PanicsWithValue(t, "bad group implementation code: this code path should be unreachable", func() {
			_ = group.Stop()
		})
	})

	t.Run("running", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		fanIn, _ := newErrorsFanIn()
		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A")
		hooks.EXPECT().OnStop("A")
		errTest := errors.New("test error")
		serviceA.EXPECT().Stop().Return(errTest)
		hooks.EXPECT().OnStopped("A", errTest)

		group := Group{
			services:        []Service{serviceA},
			fanIn:           fanIn,
			state:           StateRunning,
			hooks:           hooks,
			interceptStop:   make(chan struct{}),
			interceptDone:   make(chan struct{}),
			runningServices: map[string]struct{}{"A": {}},
		}

		// Simulate interceptRunError exiting from stop signal.
		go func() {
			<-group.interceptStop
			close(group.interceptDone)
		}()

		err := group.Stop()
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "stopping A: test error")
	})

	t.Run("already crashed", func(t *testing.T) {
		t.Parallel()

		group := Group{
			state:         StateCrashed,
			interceptDone: make(chan struct{}),
		}
		close(group.interceptDone)

		err := group.Stop()
		assert.NoError(t, err)
	})
}

func Test_Group_stop(t *testing.T) {
	t.Parallel()

	t.Run("stop two running services successfully", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		fanIn, _ := newErrorsFanIn()
		hooks := NewMockHooks(ctrl)
		serviceA := NewMockService(ctrl)
		serviceB := NewMockService(ctrl)
		serviceC := NewMockService(ctrl)

		serviceA.EXPECT().String().Return("A")
		hooks.EXPECT().OnStop("A")
		serviceA.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("A", nil)

		serviceB.EXPECT().String().Return("B")
		hooks.EXPECT().OnStop("B")
		serviceB.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("B", nil)

		serviceC.EXPECT().String().Return("C")

		group := &Group{
			services:        []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{"A": {}, "B": {}},
		}

		err := group.stop()

		assert.NoError(t, err)
		expectedGroup := &Group{
			services:        []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{},
		}
		assert.Equal(t, expectedGroup, group)
	})

	t.Run("fail to stop second of three services", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		fanIn, _ := newErrorsFanIn()
		hooks := NewMockHooks(ctrl)
		serviceA := NewMockService(ctrl)
		serviceB := NewMockService(ctrl)
		serviceC := NewMockService(ctrl)

		serviceA.EXPECT().String().Return("A")
		hooks.EXPECT().OnStop("A")
		serviceA.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("A", nil)

		serviceB.EXPECT().String().Return("B")
		hooks.EXPECT().OnStop("B")
		errTest := errors.New("test error")
		serviceB.EXPECT().Stop().Return(errTest)
		hooks.EXPECT().OnStopped("B", errTest)

		serviceC.EXPECT().String().Return("C")
		hooks.EXPECT().OnStop("C")
		serviceC.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("C", nil)

		group := &Group{
			services:        []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{"A": {}, "B": {}, "C": {}},
		}

		err := group.stop()

		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "stopping B: test error")

		expectedGroup := &Group{
			services:        []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{},
		}
		assert.Equal(t, expectedGroup, group)
	})
}
