package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readMiddlewareLog() (settings settings.MiddlewareLog, err error) {
	settings.Enabled, err = r.reader.BoolPtr("MIDDLEWARE_LOG_ENABLED")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_ENABLED: %w", err)
	}

	settings.DirPath = r.reader.String("MIDDLEWARE_LOG_DIRECTORY")

	settings.LogRequests, err = r.reader.BoolPtr("MIDDLEWARE_LOG_REQUESTS")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_REQUESTS: %w", err)
	}

	settings.LogResponses, err = r.reader.BoolPtr("MIDDLEWARE_LOG_RESPONSES")
	if err != nil {
		return settings, fmt.Errorf("environment variable MIDDLEWARE_LOG_RESPONSES: %w", err)
	}

	return settings, nil
}
