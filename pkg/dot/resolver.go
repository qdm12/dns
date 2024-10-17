package dot

import (
	"fmt"
	"net"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(settings ResolverSettings) (
	resolver *net.Resolver, err error,
) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         newDoTDial(settings),
	}, nil
}
