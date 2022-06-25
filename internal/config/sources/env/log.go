package env

import (
	"fmt"
	"os"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/log"
)

func readLog() (settings settings.Log, err error) {
	settings.Level, err = readLogLevel()
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	return settings, nil
}

func readLogLevel() (level *log.Level, err error) {
	levelString := os.Getenv("LOG_LEVEL")
	if levelString == "" {
		return nil, nil //nolint:nilnil
	}

	levelValue, err := log.ParseLevel(levelString)
	if err != nil {
		return nil, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	return &levelValue, nil
}
