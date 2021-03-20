package dot

import (
	"net"
)

// NewResolver creates a DNS over TLS resolver.
func NewResolver(options ...Option) *net.Resolver {
	settings := defaultSettings()
	for _, option := range options {
		option(&settings)
	}
	return &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial:         newDoTDial(settings),
	}
}
