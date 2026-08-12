package local

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Checker checks if a fully qualified domain name (FQDN) is local or not.
type Checker struct {
	// publicNames is a set of domain names that are considered local.
	// The trailing FQDN dot is removed for each.
	publicNames map[string]struct{}
}

func New(publicNamesAsLocal []string) *Checker {
	publicNames := make(map[string]struct{}, len(publicNamesAsLocal))
	for _, name := range publicNamesAsLocal {
		name = strings.TrimSuffix(strings.ToLower(name), ".")
		publicNames[name] = struct{}{}
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

	_, publicNameIsLocal := f.publicNames[domainName]
	if publicNameIsLocal {
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
