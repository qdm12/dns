package check

import (
	"net"
	"time"

	"github.com/qdm12/dns/internal/settings/defaults"
)

type Settings struct {
	// Resolver to use for the check.
	// It defaults to the default Go resolver.
	Resolver *net.Resolver
	// HostToResolve is the host to resolve for the check.
	// It defaults to github.com and cannot be the empty string.
	HostToResolve string
	// MaxTries is the maximum number of tries
	// before returning an error.
	// It defaults to 10 and cannot be 0.
	MaxTries int
	// WaitTime is the duration to wait between
	// each failed try. It defaults to 300ms
	// and cannot be nil in the internal state.
	WaitTime *time.Duration
	// AddWaitTime is the duration to add to the wait
	// time after each failed try.
	// It defaults to 100ms and cannot be nil
	// in the internal state.
	AddWaitTime *time.Duration
}

func (s *Settings) SetDefaults() {
	s.Resolver = defaults.Resolver(s.Resolver, net.DefaultResolver)
	s.HostToResolve = defaults.String(s.HostToResolve, "github.com")

	const defaultMaxTries = 10
	s.MaxTries = defaults.Int(s.MaxTries, defaultMaxTries)

	const defaultWaitTime = 300 * time.Millisecond
	s.WaitTime = defaults.DurationPtr(s.WaitTime, defaultWaitTime)

	const defaultAddWaitTime = 100 * time.Millisecond
	s.AddWaitTime = defaults.DurationPtr(s.AddWaitTime, defaultAddWaitTime)
}

func (s Settings) Validate() (err error) {
	return nil
}
