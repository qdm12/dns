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

func logMiddleware(userSettings settings.MiddlewareLog) (middleware *log.Middleware, err error) {
	if !*userSettings.Enabled {
		settings := log.Settings{
			Logger: noop.New(),
		}
		return log.New(settings), nil
	}

	const dirPerm = os.FileMode(0744)
	err = os.MkdirAll(userSettings.DirPath, dirPerm)
	if err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// TODO rotate log files
	const perm = os.FileMode(0644)
	filePath := filepath.Join(userSettings.DirPath, "dns.log")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return nil, err
	}

	middlewareLoggerSettings := console.Settings{
		Writer:       file,
		LogRequests:  boolPtr(*userSettings.LogRequests),
		LogResponses: boolPtr(*userSettings.LogResponses),
	}
	middlewareLogger := console.New(middlewareLoggerSettings)
	settings := log.Settings{
		Logger: middlewareLogger,
	}

	return log.New(settings), nil
}

func boolPtr(b bool) *bool { return &b }
