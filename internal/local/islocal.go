package local

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

func IsFQDNLocal(fqdn string) bool {
	if fqdn == "" {
		// Bad question really, but consider it as
		// non-local and let the upstream resolver
		// handle it.
		return false
	}

	domainName := fqdn[:len(fqdn)-1] // remove the trailing dot
	hasDot := false
	for _, c := range domainName {
		if c == '.' {
			hasDot = true
			break
		}
	}

	if !hasDot {
		// for example "localhost" or "portainer"
		return true
	}

	commonLocalTLDs := []string{
		".local",
		".lan",
		".private",
		".internal",
		".corp",
		".home",
		".network",
		".intranet",
		".site",
	}
	for _, commonLocalTLD := range commonLocalTLDs {
		if strings.HasSuffix(domainName, commonLocalTLD) {
			return true
		}
	}

	publicSuffix, icannManaged := publicsuffix.PublicSuffix(domainName)
	if icannManaged {
		return false
	} else if strings.IndexByte(publicSuffix, '.') >= 0 {
		// privately managed, such as x.y.org
		return false
	}

	return true
}
