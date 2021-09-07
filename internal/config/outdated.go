package config

import (
	"strings"

	"github.com/qdm12/golibs/params"
)

func (reader *Reader) checkOutdatedVariables() {
	outdatedToNew := map[string][]string{
		"LISTENINGPORT":       {"LISTENING_ADDRESS"},
		"PROVIDERS":           {"DOT_PROVIDERS", "DOH_PROVIDERS", "DNS_PLAINTEXT_PROVIDERS"},
		"PROVIDER":            {"DOT_PROVIDERS", "DOH_PROVIDERS", "DNS_PLAINTEXT_PROVIDERS"},
		"CACHING":             {"CACHE_TYPE", "CACHE_MAX_ENTRIES"},
		"IPV4":                {"DOT_CONNECT_IPV6"},
		"IPV6":                {"DOT_CONNECT_IPV6"},
		"UNBLOCK":             {"ALLOWED_HOSTNAMES"},
		"PRIVATE_ADDRESS":     {"REBINDING_PROTECTION"},
		"CHECK_UNBOUND":       {"CHECK_DNS"},
		"VERBOSITY":           {"LOG_LEVEL"},
		"VERBOSITY_DETAILS":   {"LOG_LEVEL"},
		"VALIDATION_LOGLEVEL": {"LOG_LEVEL"},
	}

	for outdated, new := range outdatedToNew {
		_, err := reader.env.Get(outdated, params.Compulsory())
		if err != nil { // value is not present
			continue
		}
		reader.logger.Warn("Environment variable " + outdated +
			" is deprecated, use the following instead: " + strings.Join(new, ", "))
	}
}
