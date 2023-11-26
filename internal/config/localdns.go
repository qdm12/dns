package config

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type LocalDNS struct {
	Resolvers []netip.AddrPort
}

var (
	ErrLocalResolverAddressNotValid = errors.New("local resolver address is not valid")
	ErrLocalResolverPortIsZero      = errors.New("local resolver port is zero")
)

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
	if len(l.Resolvers) == 0 {
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
	l.Resolvers, err = reader.CSVNetipAddrPorts("MIDDLEWARE_LOCALDNS_RESOLVERS")
	if err != nil {
		return err
	}
	return nil
}
