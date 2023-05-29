package check

import (
	"net"
	"time"

	"github.com/qdm12/gosettings"
)

type Settings struct {
	// Resolver to use for the check.
	// It defaults to the default Go resolver if left unset.
	Resolver *net.Resolver
	// HostToResolve is the host to resolve for the check.
	// It defaults to github.com if left unset.
	HostToResolve string
	// MaxTries is the maximum number of tries
	// before returning an error.
	// It defaults to 10 if left unset.
	MaxTries int
	// WaitTime is the duration to wait between
	// each failed try. It defaults to 300ms
	// if left unset.
	WaitTime time.Duration
	// AddWaitTime is the duration to add to the wait
	// time after each failed try.
	// It defaults to 100ms if left unset.
	AddWaitTime time.Duration
}

func (s *Settings) SetDefaults() {
	s.Resolver = gosettings.DefaultPointerRaw(s.Resolver, net.DefaultResolver)
	s.HostToResolve = gosettings.DefaultString(s.HostToResolve, "github.com")

	const defaultMaxTries = 10
	s.MaxTries = gosettings.DefaultNumber(s.MaxTries, defaultMaxTries)

	const defaultWaitTime = 300 * time.Millisecond
	s.WaitTime = gosettings.DefaultNumber(s.WaitTime, defaultWaitTime)

	const defaultAddWaitTime = 100 * time.Millisecond
	s.AddWaitTime = gosettings.DefaultNumber(s.AddWaitTime, defaultAddWaitTime)
}

func (s Settings) Validate() (err error) {
	return nil
}
