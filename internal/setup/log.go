package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/pkg/middlewares/log"
	"github.com/qdm12/dns/v2/pkg/middlewares/log/logger/console"
	"github.com/qdm12/dns/v2/pkg/middlewares/log/logger/noop"
)

func MiddlewareLogger(settings settings.MiddlewareLog) (logMiddlewareSettings log.Settings, err error) {
	if !*settings.Enabled {
		return log.Settings{
			Logger: noop.New(),
		}, nil
	}

	const dirPerm = os.FileMode(0744)
	err = os.MkdirAll(settings.DirPath, dirPerm)
	if err != nil {
		return logMiddlewareSettings, fmt.Errorf("creating log directory: %w", err)
	}

	// TODO rotate log files
	const perm = os.FileMode(0644)
	filePath := filepath.Join(settings.DirPath, "dns.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return logMiddlewareSettings, err
	}

	middlewareLoggerSettings := console.Settings{
		Writer:       file,
		LogRequests:  boolPtr(*settings.LogRequests),
		LogResponses: boolPtr(*settings.LogResponses),
	}
	middlewareLogger := console.New(middlewareLoggerSettings)
	return log.Settings{
		Logger: middlewareLogger,
	}, nil
}

func boolPtr(b bool) *bool { return &b }
