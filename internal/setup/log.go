package setup

import (
	"github.com/qdm12/dns/v2/internal/config/settings"
	pkglog "github.com/qdm12/dns/v2/pkg/log"
)

func makeMiddlewareLogger(logger Logger,
	userSettings settings.Log) (mLogger *MiddlewareLogger) {
	return &MiddlewareLogger{
		logger:             logger,
		logRequest:         *userSettings.LogRequests,
		logResponse:        *userSettings.LogResponses,
		logRequestResponse: *userSettings.LogRequestsResponses,
	}
}

// TODO config in middleware pkg.
type MiddlewareLogger struct {
	logger             pkglog.Logger
	logRequest         bool
	logResponse        bool
	logRequestResponse bool
}

func (m *MiddlewareLogger) Error(s string) { m.logger.Error(s) }
func (m *MiddlewareLogger) LogRequest(s string) {
	if m.logRequest {
		m.logger.Info(s)
	}
}
func (m *MiddlewareLogger) LogResponse(s string) {
	if m.logResponse {
		m.logger.Info(s)
	}
}
func (m *MiddlewareLogger) LogRequestResponse(s string) {
	if m.logRequestResponse {
		m.logger.Info(s)
	}
}
