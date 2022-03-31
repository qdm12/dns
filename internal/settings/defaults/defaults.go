package defaults

import (
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

func IP(existing, defaultValue net.IP) net.IP {
	if existing != nil {
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

func PrometheusRegisterer(existing, //nolint:ireturn
	defaultValue prometheus.Registerer) (
	result prometheus.Registerer,
) {
	if existing != nil {
		return existing
	}
	return defaultValue
}
