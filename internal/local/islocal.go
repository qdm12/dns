package local

import (
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
)

// Checker checks if a fully qualified domain name (FQDN) is local or not.
type Checker struct {
	// publicNames is a list of domain names that are considered local.
	// The trailing FQDN dot is removed for each.
	// Note a public name is considered local if its parent domain is contained in this list.
	publicNames []string
}

func New(publicNamesAsLocal []string) *Checker {
	publicNames := make([]string, len(publicNamesAsLocal))
	for i, name := range publicNamesAsLocal {
		publicNames[i] = strings.TrimSuffix(strings.ToLower(name), ".")
	}

	return &Checker{
		publicNames: publicNames,
	}
}

func (f *Checker) IsFQDNLocal(fqdn string) bool {
	if fqdn == "" {
		// Bad question really, but consider it as
		// non-local and let the upstream resolver
		// handle it.
		return false
	}

	domainName := fqdn[:len(fqdn)-1] // remove the trailing dot
	domainName = strings.ToLower(domainName)
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

	commonLocalTLDs := [...]string{
		".local",
		".lan",
		".private",
		".internal",
		".corp",
		".home",
		".intranet",
	}
	for _, commonLocalTLD := range commonLocalTLDs {
		if strings.HasSuffix(domainName, commonLocalTLD) {
			return true
		}
	}

	if f.treatDomainAsLocal(domainName) {
		return true
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

func (f *Checker) treatDomainAsLocal(domainName string) bool {
	for _, publicName := range f.publicNames {
		if dns.IsSubDomain(publicName, domainName) {
			return true
		}
	}
	return false
}
