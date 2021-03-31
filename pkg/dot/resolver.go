package dot

import (
	"net"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(settings Settings) *net.Resolver {
	settings.setDefaults()
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         newDoTDial(settings),
	}
}
