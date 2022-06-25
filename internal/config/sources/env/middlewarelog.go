package env

import (
	"fmt"
	"os"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readMiddlewareLog() (settings settings.MiddlewareLog, err error) {
	settings.Enabled, err = envToBoolPtr("MIDDLEWARE_LOG_ENABLED")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_ENABLED: %w", err)
	}

	settings.DirPath = os.Getenv("MIDDLEWARE_LOG_DIRECTORY")

	settings.LogRequests, err = envToBoolPtr("MIDDLEWARE_LOG_REQUESTS")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_REQUESTS: %w", err)
	}

	settings.LogResponses, err = envToBoolPtr("MIDDLEWARE_LOG_RESPONSES")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_RESPONSES: %w", err)
	}

	return settings, nil
}
