//go:build !js && !windows

package nameserver

import (
	"fmt"
	"net/netip"
	"os"
	"strings"
)

// GetDNSServers retrieves the nameserver IP addresses from /etc/resolv.conf on Unix-like systems.
// If an error is encountered, it does still return nameservers correctly found together with the error.
func GetDNSServers() (nameservers []netip.Addr, err error) {
	const filename = "/etc/resolv.conf"
	return getLocalNameservers(filename)
}

func getLocalNameservers(filename string) (nameservers []netip.Addr, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var errs []error
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "nameserver" {
			continue
		}
		for _, field := range fields[1:] {
			ip, err := netip.ParseAddr(field)
			if err != nil {
				errs = append(errs, fmt.Errorf("parsing nameserver address: %w", err))
				continue
			}
			nameservers = append(nameservers, ip)
		}
	}

	if len(errs) > 0 {
		err = joinErrs(errs)
	}

	return nameservers, err
}

func joinErrs(errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	wVerbs := make([]string, len(errs))
	for i := range wVerbs {
		wVerbs[i] = "%w"
	}
	format := strings.Join(wVerbs, "; ")

	args := make([]any, len(errs))
	for i := range errs {
		args[i] = errs[i]
	}
	return fmt.Errorf(format, args...)
}
