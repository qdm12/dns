package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readMetrics() (settings settings.Metrics, err error) {
	settings.Type = strings.ToLower(os.Getenv("METRICS_TYPE"))
	if err != nil {
		return settings, fmt.Errorf("environment variable METRICS_TYPE: %w", err)
	}

	settings.Prometheus = readPrometheus()

	return settings, nil
}
