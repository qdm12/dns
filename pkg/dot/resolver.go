package dot

import (
	"net"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(settings ResolverSettings) (
	resolver *net.Resolver, err error) {
	settings.SetDefaults()

	dial, err := newDoTDial(settings)
	if err != nil {
		return nil, err
	}

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         dial,
	}, nil
}
