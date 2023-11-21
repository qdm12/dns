package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/log"
)

func (r *Reader) readLog() (settings settings.Log, err error) {
	settings.Level, err = r.readLogLevel()
	if err != nil {
		return settings, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	return settings, nil
}

func (r *Reader) readLogLevel() (level *log.Level, err error) {
	levelString := r.reader.String("LOG_LEVEL")
	if levelString == "" {
		return nil, nil //nolint:nilnil
	}

	levelValue, err := log.ParseLevel(levelString)
	if err != nil {
		return nil, fmt.Errorf("environment variable LOG_LEVEL: %w", err)
	}

	return &levelValue, nil
}
