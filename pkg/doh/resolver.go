package doh

import (
	"fmt"
	"net"
)

// NewResolver creates a DNS over HTTPs resolver.
func NewResolver(settings ResolverSettings) (
	resolver *net.Resolver, err error) {
	settings.SetDefaults()

	dial, err := newDoHDial(settings)
	if err != nil {
		return nil, fmt.Errorf("creating DoH dial: %w", err)
	}

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         dial,
	}, nil
}
