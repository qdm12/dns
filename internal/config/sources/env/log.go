package env

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/golibs/logging"
)

func readLog() (settings settings.Log, err error) {
	settings.Level, err = readLogLevel()
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	settings.LogRequests, err = envToBoolPtr("LOG_REQUESTS")
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_REQUESTS: %w", err)
	}

	settings.LogResponses, err = envToBoolPtr("LOG_RESPONSES")
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_RESPONSES: %w", err)
	}

	settings.LogRequestsResponses, err = envToBoolPtr("LOG_REQUESTS_RESPONSES")
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_REQUESTS_RESPONSES: %w", err)
	}

	return settings, nil
}

func loggingLevelPtr(level logging.Level) *logging.Level { return &level }

var ErrUnknownLogLevel = errors.New("unknown log level")

func readLogLevel() (level *logging.Level, err error) {
	levelString := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch levelString {
	case "":
		return nil, nil //nolint:nilnil
	case "debug":
		return loggingLevelPtr(logging.LevelDebug), nil
	case "info":
		return loggingLevelPtr(logging.LevelInfo), nil
	case "warning":
		return loggingLevelPtr(logging.LevelWarn), nil
	case "error":
		return loggingLevelPtr(logging.LevelError), nil
	default:
		return nil, fmt.Errorf("%w: %s: can be one of: debug, info, warning, error",
			ErrUnknownLogLevel, levelString)
	}
}
