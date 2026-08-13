package config

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type LocalDNS struct {
	Enabled                  *bool
	Resolvers                []netip.AddrPort
	PublicNamesAsLocal       []string
	PublicNameserversAsLocal []netip.Prefix
}

var (
	ErrLocalResolverAddressNotValid = errors.New("local resolver address is not valid")
	ErrLocalResolverPortIsZero      = errors.New("local resolver port is zero")
)

func (l *LocalDNS) setDefault() {
	l.Enabled = gosettings.DefaultPointer(l.Enabled, true)
	privateNameservers, _ := nameserver.GetPrivateDNSServersWithPublicCIDRsAsLocal(
		l.PublicNameserversAsLocal)
	l.Resolvers = gosettings.DefaultSlice(l.Resolvers, addrsToAddr53(privateNameservers))
	l.PublicNamesAsLocal = gosettings.DefaultSlice(l.PublicNamesAsLocal, []string{})
	l.PublicNameserversAsLocal = gosettings.DefaultSlice(l.PublicNameserversAsLocal,
		[]netip.Prefix{})
}

func addrsToAddr53(addrs []netip.Addr) (addrPorts []netip.AddrPort) {
	addrPorts = make([]netip.AddrPort, len(addrs))
	const dnsPort = 53
	for i := range addrs {
		addrPorts[i] = netip.AddrPortFrom(addrs[i], dnsPort)
	}
	return addrPorts
}

func (l *LocalDNS) validate() (err error) {
	for _, resolver := range l.Resolvers {
		switch {
		case !resolver.IsValid():
			return fmt.Errorf("%w: %s",
				ErrLocalResolverAddressNotValid, resolver)
		case resolver.Port() == 0:
			return fmt.Errorf("%w: %s",
				ErrLocalResolverPortIsZero, resolver)
		}
	}

	for _, prefix := range l.PublicNameserversAsLocal {
		if !prefix.IsValid() {
			return fmt.Errorf("nameserver public CIDR is not valid: %s", prefix)
		}
	}

	return nil
}

func (l *LocalDNS) String() string {
	return l.ToLinesNode().String()
}

func (l *LocalDNS) ToLinesNode() (node *gotree.Node) {
	if !*l.Enabled {
		return gotree.New("Local DNS middleware: disabled")
	}

	node = gotree.New("Local DNS middleware:")
	resolversNode := gotree.New("Local resolvers:")
	for _, resolver := range l.Resolvers {
		resolversNode.Appendf("%s", resolver)
	}
	node.AppendNode(resolversNode)

	if len(l.PublicNamesAsLocal) > 0 {
		node.Appendf("Public names considered local: %s",
			strings.Join(l.PublicNamesAsLocal, ", "))
	}

	if len(l.PublicNameserversAsLocal) > 0 {
		cidrs := make([]string, len(l.PublicNameserversAsLocal))
		for i, cidr := range l.PublicNameserversAsLocal {
			cidrs[i] = cidr.String()
		}
		node.Appendf("Public nameserver CIDRs considered local: %s", strings.Join(cidrs, ", "))
	}

	return node
}

func (l *LocalDNS) read(reader *reader.Reader) (err error) {
	l.Enabled, err = reader.BoolPtr("MIDDLEWARE_LOCALDNS_ENABLED")
	if err != nil {
		return err
	}

	l.Resolvers, err = reader.CSVNetipAddrPorts("MIDDLEWARE_LOCALDNS_RESOLVERS")
	if err != nil {
		return err
	}

	l.PublicNamesAsLocal = reader.CSV("MIDDLEWARE_LOCALDNS_PUBLIC_NAMES_AS_LOCAL")

	l.PublicNameserversAsLocal, err = reader.CSVNetipPrefixes(
		"MIDDLEWARE_LOCALDNS_NAMESERVERS_PUBLIC_CIDRS_AS_LOCAL")
	if err != nil {
		return err
	}

	return nil
}
