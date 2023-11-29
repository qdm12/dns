package config

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type LocalDNS struct {
	Enabled   *bool
	Resolvers []netip.AddrPort
}

var (
	ErrLocalResolverAddressNotValid = errors.New("local resolver address is not valid")
	ErrLocalResolverPortIsZero      = errors.New("local resolver port is zero")
)

func (l *LocalDNS) setDefault() {
	l.Enabled = gosettings.DefaultPointer(l.Enabled, true)
	l.Resolvers = gosettings.DefaultSlice(l.Resolvers,
		nameserver.GetDNSServers())
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
	return nil
}
