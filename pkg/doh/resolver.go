package doh

import (
	"fmt"
	"net"
)

// NewResolver creates a DNS over HTTPs resolver.
func NewResolver(settings ResolverSettings) (
	resolver *net.Resolver, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	dial := newDoHDial(settings)

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         dial,
	}, nil
}
