package services

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/internal/services/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSequence(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	dummyService := NewMockService(ctrl)
	dummyService.EXPECT().String().Return("A").AnyTimes()

	testCases := map[string]struct {
		settings    SequenceSettings
		sequence    *Sequence
		errSentinel error
		errMessage  string
	}{
		"empty settings": {
			errSentinel: ErrNoServiceStart,
			errMessage:  "validating settings: no service start order specified",
		},
		"full settings": {
			settings: SequenceSettings{
				Name:          "name",
				ServicesStart: []Service{dummyService},
				ServicesStop:  []Service{dummyService},
				Hooks:         hooks.NewWithLog(nil),
			},
			sequence: &Sequence{
				name:            "name",
				servicesStart:   []Service{dummyService},
				servicesStop:    []Service{dummyService},
				hooks:           hooks.NewWithLog(nil),
				runningServices: map[string]struct{}{},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sequence, err := NewSequence(testCase.settings)

			assert.ErrorIs(t, err, testCase.errSentinel)
			if testCase.errSentinel != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.sequence, sequence)
		})
	}
}

func Test_Sequence_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		sequence *Sequence
		expected string
	}{
		"empty name": {
			sequence: &Sequence{},
			expected: "sequence",
		},
		"set name": {
			sequence: &Sequence{
				name: "A",
			},
			expected: "sequence A",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := testCase.sequence.String()

			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func Test_Sequence_Start(t *testing.T) {
	t.Parallel()

	errTest := errors.New("test error")

	t.Run("panic if already running", func(t *testing.T) {
		t.Parallel()

		sequence := &Sequence{
			name:  "name",
			state: StateRunning,
		}

		assert.PanicsWithValue(t,
			"sequence name already running",
			func() {
				_, _ = sequence.Start()
			})
	})

	t.Run("first service start error", func(t *testing.T) {
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
		serviceB.EXPECT().String().Return("B")          // stop method

		settings := SequenceSettings{
			ServicesStart: []Service{serviceA, serviceB},
			ServicesStop:  []Service{serviceA, serviceB},
			Hooks:         hooks,
		}

		sequence, err := NewSequence(settings)
		require.NoError(t, err)

		runError, err := sequence.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "starting A: test error")
	})

	t.Run("second service start error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		hooks := NewMockHooks(ctrl)

		serviceA := NewMockService(ctrl)
		serviceA.EXPECT().String().Return("A").Times(2) // settings validation
		hooks.EXPECT().OnStart("A")
		serviceA.EXPECT().String().Return("A") // start method
		runErrorA := make(chan error)
		serviceA.EXPECT().Start().Return(runErrorA, nil)
		hooks.EXPECT().OnStarted("A", nil)

		serviceB := NewMockService(ctrl)
		serviceB.EXPECT().String().Return("B").Times(2) // settings validation
		hooks.EXPECT().OnStart("B")
		serviceB.EXPECT().String().Return("B") // start method
		serviceB.EXPECT().Start().Return(nil, errTest)
		hooks.EXPECT().OnStarted("B", errTest)

		serviceB.EXPECT().String().Return("B") // stop method
		serviceA.EXPECT().String().Return("A") // stop method
		hooks.EXPECT().OnStop("A")
		serviceA.EXPECT().Stop().Return(nil) // ignored error
		hooks.EXPECT().OnStopped("A", nil)

		settings := SequenceSettings{
			ServicesStart: []Service{serviceA, serviceB},
			ServicesStop:  []Service{serviceA, serviceB},
			Hooks:         hooks,
		}

		sequence, err := NewSequence(settings)
		require.NoError(t, err)

		runError, err := sequence.Start()

		assert.Nil(t, runError)
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "starting B: test error")
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

		settings := SequenceSettings{
			ServicesStart: []Service{serviceA, serviceB},
			ServicesStop:  []Service{serviceB, serviceA},
			Hooks:         hooks,
		}

		sequence, err := NewSequence(settings)
		require.NoError(t, err)

		runError, err := sequence.Start()

		require.NoError(t, err)
		require.NotNil(t, runError)

		select {
		case err := <-runError:
			assert.NoError(t, err)
		default:
		}

		// Expectations for the sequence stop call.
		serviceA.EXPECT().String().Return("A") // stop method
		serviceB.EXPECT().String().Return("B") // stop method
		hooks.EXPECT().OnStop("B")
		serviceB.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("B", nil)
		hooks.EXPECT().OnStop("A")
		serviceA.EXPECT().Stop().Return(nil)
		hooks.EXPECT().OnStopped("A", nil)

		err = sequence.Stop()
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

		settings := SequenceSettings{
			ServicesStart: []Service{serviceA, serviceB},
			ServicesStop:  []Service{serviceB, serviceA},
			Hooks:         hooks,
		}

		sequence, err := NewSequence(settings)
		require.NoError(t, err)

		runError, startErr := sequence.Start()
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

func Test_Sequence_interceptRunError(t *testing.T) {
	t.Parallel()

	t.Run("stop", func(t *testing.T) {
		t.Parallel()

		sequence := Sequence{
			interceptStop: make(chan struct{}),
			interceptDone: make(chan struct{}),
		}
		ready := make(chan struct{})
		output := make(chan error)
		go sequence.interceptRunError(ready, nil, output)
		<-ready
		close(sequence.interceptStop)
		<-sequence.interceptDone
	})

	t.Run("service crash and sequence is already stopping", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		service := NewMockService(ctrl)

		sequence := &Sequence{
			servicesStop:  []Service{service},
			state:         StateStopping,
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		input := make(chan serviceError)
		output := make(chan error)
		close(output) // do not write to output channel

		go sequence.interceptRunError(ready, input, output)

		<-ready

		errTest := errors.New("test error")
		input <- serviceError{
			format:      errorFormatCrash,
			serviceName: "A",
			err:         errTest,
		}

		<-sequence.interceptDone

		expectedSequence := &Sequence{
			servicesStop: []Service{service},
			state:        StateStopping,
		}

		_, ok := <-sequence.interceptDone
		assert.False(t, ok)
		sequence.interceptDone = nil

		assert.Equal(t, expectedSequence, sequence)
	})

	t.Run("one of two services crash", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		fanIn, _ := newErrorsFanIn()
		hooks := NewMockHooks(ctrl)
		serviceA := NewMockService(ctrl)
		serviceB := NewMockService(ctrl)

		// Expectations for stop method call.
		serviceA.EXPECT().String().Return("A")
		serviceB.EXPECT().String().Return("B")
		hooks.EXPECT().OnStop("B")
		errStop := errors.New("stop error")
		serviceB.EXPECT().Stop().Return(errStop) // ignored error
		hooks.EXPECT().OnStopped("B", errStop)

		sequence := &Sequence{
			runningServices: map[string]struct{}{
				"A": {},
				"B": {},
			},
			servicesStop:  []Service{serviceA, serviceB},
			fanIn:         fanIn,
			hooks:         hooks,
			state:         StateRunning,
			interceptStop: make(chan struct{}),
			interceptDone: make(chan struct{}),
		}

		ready := make(chan struct{})
		input := make(chan serviceError)
		output := make(chan error)

		go sequence.interceptRunError(ready, input, output)

		<-ready

		errTest := errors.New("test error")
		hooks.EXPECT().OnCrash("A", errTest)
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

		<-sequence.interceptDone

		expectedSequence := &Sequence{
			runningServices: map[string]struct{}{},
			servicesStop:    []Service{serviceA, serviceB},
			fanIn:           fanIn,
			hooks:           hooks,
			state:           StateCrashed,
		}

		_, ok = <-sequence.interceptDone
		assert.False(t, ok)
		sequence.interceptDone = nil

		close(sequence.interceptStop)
		sequence.interceptStop = nil

		assert.Equal(t, expectedSequence, sequence)
	})
}

func Test_Sequence_Stop(t *testing.T) {
	t.Parallel()

	t.Run("already crashed", func(t *testing.T) {
		t.Parallel()

		sequence := Sequence{
			name:          "name",
			state:         StateCrashed,
			interceptDone: make(chan struct{}),
		}
		close(sequence.interceptDone)

		err := sequence.Stop()
		require.NoError(t, err)
	})

	t.Run("already stopped", func(t *testing.T) {
		t.Parallel()

		sequence := Sequence{
			name:  "name",
			state: StateStopped,
		}
		assert.PanicsWithValue(t, "bad calling code: sequence name already stopped", func() {
			_ = sequence.Stop()
		})
	})

	t.Run("in starting state", func(t *testing.T) {
		t.Parallel()

		sequence := Sequence{
			name:  "name",
			state: StateStarting,
		}
		assert.PanicsWithValue(t,
			"bad sequence implementation code: this code path should be unreachable",
			func() {
				_ = sequence.Stop()
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

		sequence := Sequence{
			servicesStop:    []Service{serviceA},
			fanIn:           fanIn,
			state:           StateRunning,
			hooks:           hooks,
			interceptStop:   make(chan struct{}),
			interceptDone:   make(chan struct{}),
			runningServices: map[string]struct{}{"A": {}},
		}

		// Simulate interceptRunError exiting from stop signal.
		go func() {
			<-sequence.interceptStop
			close(sequence.interceptDone)
		}()

		err := sequence.Stop()
		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "stopping A: test error")
	})
}

func Test_Sequence_stop(t *testing.T) {
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

		sequence := &Sequence{
			servicesStop:    []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{"A": {}, "B": {}},
		}

		err := sequence.stop()

		assert.NoError(t, err)
		expectedSequence := &Sequence{
			servicesStop:    []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{},
		}
		assert.Equal(t, expectedSequence, sequence)
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

		sequence := &Sequence{
			servicesStop:    []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{"A": {}, "B": {}, "C": {}},
		}

		err := sequence.stop()

		assert.ErrorIs(t, err, errTest)
		assert.EqualError(t, err, "stopping B: test error")

		expectedSequence := &Sequence{
			servicesStop:    []Service{serviceA, serviceB, serviceC},
			fanIn:           fanIn,
			hooks:           hooks,
			runningServices: map[string]struct{}{},
		}
		assert.Equal(t, expectedSequence, sequence)
	})
}
