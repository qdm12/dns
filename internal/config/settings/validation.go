package settings

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/qdm12/dns/v2/pkg/provider"
)

var ErrHostnameNotValid = errors.New("hostname is not valid")

var regexHostname = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9_])(\.([a-zA-Z0-9]|[a-zA-Z0-9_][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]))*$`) //nolint:lll

func checkHostnames(hostnames []string) (err error) {
	for _, hostname := range hostnames {
		if !regexHostname.MatchString(hostname) {
			return fmt.Errorf("%w: %s", ErrHostnameNotValid, hostname)
		}
	}
	return nil
}

var ErrValueNotOneOf = errors.New("value is not one of the accepted values")

func checkIsOneOf(value string, acceptedValues ...string) (err error) {
	for _, acceptedValue := range acceptedValues {
		if value == acceptedValue {
			return nil
		}
	}
	return fmt.Errorf("%w: %q must be one of: %s",
		ErrValueNotOneOf, value, strings.Join(acceptedValues, ", "))
}

func checkListeningAddress(address string) (err error) {
	_, _, err = net.SplitHostPort(address)
	return err
}

func checkProviderNames(providerNames []string) (err error) {
	allProviders := provider.All()
	allProviderNames := make([]string, len(allProviders))
	for i, provider := range allProviders {
		allProviderNames[i] = provider.Name
	}

	for _, providerName := range providerNames {
		valid := false
		for _, acceptedProviderName := range allProviderNames {
			if strings.EqualFold(providerName, acceptedProviderName) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("%w: %q must be one of: %s",
				ErrValueNotOneOf, providerName, orStrings(allProviderNames))
		}
	}

	return nil
}
