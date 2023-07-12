package env

import (
	"fmt"
	"os"
	"strings"
)

func checkOutdatedVariables() (warnings []string) {
	outdatedToNew := map[string][]string{
		"LISTENINGPORT":       {"LISTENING_ADDRESS"},
		"PROVIDERS":           {"DOT_RESOLVERS", "DOH_RESOLVERS", "DNS_FALLBACK_PLAINTEXT_RESOLVERS"},
		"PROVIDER":            {"DOT_RESOLVERS", "DOH_RESOLVERS", "DNS_FALLBACK_PLAINTEXT_RESOLVERS"},
		"CACHING":             {"CACHE_TYPE"},
		"IPV4":                {"DOT_CONNECT_IPV6"},
		"IPV6":                {"DOT_CONNECT_IPV6"},
		"UNBLOCK":             {"ALLOWED_HOSTNAMES"},
		"PRIVATE_ADDRESS":     {"REBINDING_PROTECTION"},
		"CHECK_UNBOUND":       {"CHECK_DNS"},
		"VERBOSITY":           {"LOG_LEVEL"},
		"VERBOSITY_DETAILS":   {"LOG_LEVEL", "MIDDLEWARE_LOG_ENABLED", "MIDDLEWARE_LOG_DIRECTORY", "MIDDLEWARE_LOG_REQUESTS", "MIDDLEWARE_LOG_RESPONSES"}, //nolint:lll
		"VALIDATION_LOGLEVEL": {"LOG_LEVEL", "MIDDLEWARE_LOG_ENABLED", "MIDDLEWARE_LOG_DIRECTORY", "MIDDLEWARE_LOG_REQUESTS", "MIDDLEWARE_LOG_RESPONSES"}, //nolint:lll
	}

	for outdated, new := range outdatedToNew {
		value := os.Getenv(outdated)
		if value == "" {
			continue
		}

		replacementVariables := strings.Join(new, ", ")
		warning := fmt.Sprintf("Environment variable %s is no longer functional, "+
			"use the following instead: %s", outdated, replacementVariables)
		warnings = append(warnings, warning)
	}

	return warnings
}
