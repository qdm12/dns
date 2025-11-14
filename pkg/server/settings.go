package server

import (
	"errors"
	"fmt"
	"os"

	lognoop "github.com/qdm12/dns/v2/pkg/log/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// ListeningAddress is the server listening address, and defaults
	// to ":53".
	ListeningAddress *string
	// Dialer is used to establish connections with upstream resolvers
	// and must be set.
	Dialer Dialer
	// Middlewares is a list of middlewares to use.
	// The first one is the first wrapper, and the last one
	// is the last wrapper of the handlers in the chain.
	Middlewares []Middleware
	// Logger is the logger to log information.
	// It defaults to a No-Op logger implementation.
	Logger Logger
	// TimeoutWarn indicates whether to log timeout errors at the
	// warning level or at the debug level. It defaults to false.
	TimeoutWarn *bool
}

func (s *Settings) SetDefaults() {
	s.ListeningAddress = gosettings.DefaultPointer(s.ListeningAddress, ":53")
	s.Logger = gosettings.DefaultComparable[Logger](s.Logger, lognoop.New())
	s.TimeoutWarn = gosettings.DefaultPointer(s.TimeoutWarn, false)
}

var (
	ErrListeningAddressNotValid = errors.New("listening address is not valid")
	ErrDialerNotSet             = errors.New("dialer is not set")
)

func (s Settings) Validate() (err error) {
	err = validate.ListeningAddress(*s.ListeningAddress, os.Getuid())
	if err != nil {
		return fmt.Errorf("%w: %s", ErrListeningAddressNotValid, *s.ListeningAddress)
	}

	if s.Dialer == nil {
		return fmt.Errorf("%w", ErrDialerNotSet)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Server settings:")
	node.Appendf("Listening address: %s", *s.ListeningAddress)
	node.Appendf("Upstream resolver connection type: %s", s.Dialer)
	node.Appendf("Log timeout at the warning level: %s", gosettings.BoolToYesNo(s.TimeoutWarn))

	if len(s.Middlewares) > 0 {
		middlewares := node.Append("Middlewares:")
		for _, middleware := range s.Middlewares {
			middlewares.Append(middleware.String())
		}
	}

	return node
}
