package dot

import (
	"net"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(settings ResolverSettings) *net.Resolver {
	settings.SetDefaults()
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         newDoTDial(settings),
	}
}
