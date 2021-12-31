package doh

import (
	"net"
)

// NewResolver creates a DNS over HTTPs resolver.
func NewResolver(settings ResolverSettings) *net.Resolver {
	settings.SetDefaults()
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         newDoHDial(settings),
	}
}
