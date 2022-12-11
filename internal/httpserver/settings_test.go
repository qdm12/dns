package httpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Settings_SetDefaults(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings         Settings
		expectedSettings Settings
	}{
		"empty settings": {
			expectedSettings: Settings{
				Name:              stringPtr(""),
				Address:           stringPtr(""),
				ShutdownTimeout:   3 * time.Second,
				ReadTimeout:       10 * time.Second,
				ReadHeaderTimeout: time.Second,
				Logger:            &noopLogger{},
			},
		},
		"all settings fields set": {
			settings: Settings{
				Name:              stringPtr("x"),
				Handler:           http.NewServeMux(),
				Address:           stringPtr("test"),
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
				Logger:            NewMockInfoer(nil),
			},
			expectedSettings: Settings{
				Name:              stringPtr("x"),
				Handler:           http.NewServeMux(),
				Address:           stringPtr("test"),
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: 2 * time.Second,
				ShutdownTimeout:   3 * time.Second,
				Logger:            NewMockInfoer(nil),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.settings.SetDefaults()

			assert.Equal(t, testCase.expectedSettings, testCase.settings)
		})
	}
}

func Test_Settings_Validate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings   Settings
		errMessage string
	}{
		"nil handler": {
			errMessage: "handler is nil",
		},
		"invalid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
				Address: stringPtr(":-1"),
			},
			errMessage: "listening address is not valid: address -1: invalid port",
		},
		"valid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
				Address: stringPtr(":0"),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.Validate()

			if testCase.errMessage == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testCase.errMessage)
			}
		})
	}
}
