package defaults

import (
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/log"
)

func String(existing, defaultValue string) string {
	if existing != "" {
		return existing
	}
	return defaultValue
}

func Int(existing, defaultValue int) int {
	if existing != 0 {
		return existing
	}
	return defaultValue
}

func Uint16(existing, defaultValue uint16) uint16 {
	if existing != 0 {
		return existing
	}
	return defaultValue
}

func Duration(existing time.Duration, defaultValue time.Duration) time.Duration {
	if existing != 0 {
		return existing
	}
	return defaultValue
}

func IP(existing, defaultValue netip.Addr) netip.Addr {
	if existing.IsValid() {
		return existing
	}
	return defaultValue
}

func BoolPtr(existing *bool, defaultValue bool) *bool {
	if existing != nil {
		return existing
	}
	return &defaultValue
}

func StringPtr(existing *string, defaultValue string) *string {
	if existing != nil {
		return existing
	}
	return &defaultValue
}

func DurationPtr(existing *time.Duration, defaultValue time.Duration) *time.Duration {
	if existing != nil {
		return existing
	}
	return &defaultValue
}

func LogLevelPtr(existing *log.Level, defaultValue log.Level) *log.Level {
	if existing != nil {
		return existing
	}
	return &defaultValue
}

func HTTPClient(existing, defaultValue *http.Client) *http.Client {
	if existing != nil {
		return existing
	}
	return defaultValue
}

func Resolver(existing, defaultValue *net.Resolver) *net.Resolver {
	if existing != nil {
		return existing
	}
	return defaultValue
}

func PrometheusRegisterer(existing,
	defaultValue prometheus.Registerer) (
	result prometheus.Registerer) {
	if existing != nil {
		return existing
	}
	return defaultValue
}
