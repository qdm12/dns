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

func allProvidersStringSet() (set map[string]struct{}) {
	providers := provider.All()
	set = make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		set[provider.Name] = struct{}{}
	}
	return set
}
