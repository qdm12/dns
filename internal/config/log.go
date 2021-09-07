package config

import (
	"fmt"

	pkglog "github.com/qdm12/dns/pkg/log"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
)

func (settings *Settings) PatchLogger(logger pkglog.Logger) {
	settings.DoT.Logger = logger
	settings.DoH.Logger = logger
	middlewareLogger := settings.Log.makeMiddlewareLogger(logger)
	settings.DoT.LogMiddleware.Logger = middlewareLogger
	settings.DoH.LogMiddleware.Logger = middlewareLogger
}

type Log struct {
	Level                logging.Level
	LogRequests          bool
	LogResponses         bool
	LogRequestsResponses bool
}

func getLogSettings(env params.Interface) (settings Log, err error) {
	settings.Level, err = env.LogLevel("LOG_LEVEL", params.Default("info"))
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	settings.LogRequests, err = env.OnOff("LOG_REQUESTS", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_REQUESTS: %w", err)
	}

	settings.LogResponses, err = env.OnOff("LOG_RESPONSES", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_RESPONSES: %w", err)
	}

	settings.LogRequestsResponses, err = env.OnOff("LOG_REQUESTS_RESPONSES", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_REQUESTS_RESPONSES: %w", err)
	}

	return settings, nil
}

func (l *Log) makeMiddlewareLogger(logger pkglog.Logger) *middlewareLogger {
	return &middlewareLogger{
		logger:             logger,
		logRequest:         l.LogRequests,
		logResponse:        l.LogResponses,
		logRequestResponse: l.LogRequestsResponses,
	}
}

type middlewareLogger struct {
	logger             pkglog.Logger
	logRequest         bool
	logResponse        bool
	logRequestResponse bool
}

func (m *middlewareLogger) Error(s string) { m.logger.Error(s) }
func (m *middlewareLogger) LogRequest(s string) {
	if m.logRequest {
		m.logger.Info(s)
	}
}
func (m *middlewareLogger) LogResponse(s string) {
	if m.logResponse {
		m.logger.Info(s)
	}
}
func (m *middlewareLogger) LogRequestResponse(s string) {
	if m.logRequestResponse {
		m.logger.Info(s)
	}
}

func (l *Log) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Level: "+l.Level.String())

	if l.LogRequests {
		lines = append(lines, subSection+"Log requests: enabled")
	}

	if l.LogResponses {
		lines = append(lines, subSection+"Log responses: enabled")
	}

	if l.LogRequestsResponses {
		lines = append(lines, subSection+"Log resquests => responses: enabled")
	}

	return lines
}
