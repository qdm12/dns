package httpserver

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/qdm12/dns/v2/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings       Settings
		expectedServer *Server
		errMessage     string
	}{
		"invalid settings": {
			errMessage: "validating settings: handler is nil",
		},
		"valid settings": {
			settings: Settings{
				Handler: http.NewServeMux(),
			},
			expectedServer: &Server{
				settings: Settings{
					Name:              stringPtr(""),
					Handler:           http.NewServeMux(),
					Address:           stringPtr(""),
					ShutdownTimeout:   3 * time.Second,
					ReadTimeout:       10 * time.Second,
					ReadHeaderTimeout: time.Second,
					Logger:            &noopLogger{},
				},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := New(testCase.settings)

			if testCase.errMessage == "" {
				assert.NoError(t, err)
				assert.NotNil(t, server.service)
				server.service = nil
			} else {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.expectedServer, server)
		})
	}
}
func Test_Server_GetAddress(t *testing.T) {
	t.Parallel()

	server := &Server{
		listeningAddress: "x",
	}

	address := server.GetAddress()

	assert.Equal(t, "x", address)
}

func Test_Server_success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	logger := NewMockInfoer(ctrl)
	logger.EXPECT().Info(newRegexMatcher("^test http server listening on 127.0.0.1:[1-9][0-9]{0,4}$"))

	server := &Server{
		settings: Settings{
			Name:            stringPtr("test"),
			Address:         stringPtr("127.0.0.1:0"),
			ShutdownTimeout: 10 * time.Second,
			Logger:          logger,
		},
	}

	serverService := services.NewRunWrapper("server", server.run)

	runError, err := serverService.Start()
	require.NoError(t, err)

	addressRegex := regexp.MustCompile(`^127.0.0.1:[1-9][0-9]{0,4}$`)
	address := server.GetAddress()
	assert.Regexp(t, addressRegex, address)

	select {
	case err := <-runError:
		require.NoError(t, err)
	default:
	}

	err = serverService.Stop()
	require.NoError(t, err)
}

func Test_Server_startError(t *testing.T) {
	t.Parallel()

	server := &Server{
		settings: Settings{
			Address:         stringPtr("127.0.0.1:-1"),
			ShutdownTimeout: 10 * time.Second,
		},
	}

	serverService := services.NewRunWrapper("server", server.run)

	runtimeError, err := serverService.Start()

	require.EqualError(t, err, "listen tcp: address -1: invalid port")
	assert.Nil(t, runtimeError)
}
