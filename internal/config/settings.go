package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

// Settings contain settings to configure the entire program
// and to live patch the program from external sources.
type Settings struct {
	// Upstream is the DNS upstream connection type
	// and can be either 'dot' or 'doh'.
	// It defaults to 'dot' if left uset.
	Upstream string
	// ListeningAddress is the DNS server listening address.
	// It can be set to the empty string to listen on all interfaces
	// on a random available port.
	// It defaults to ":53".
	ListeningAddress *string
	Block            Block
	Cache            Cache
	DoH              DoH
	DoT              DoT
	Log              Log
	MiddlewareLog    MiddlewareLog
	Metrics          Metrics
	LocalDNS         LocalDNS
	CheckDNS         *bool
	UpdatePeriod     *time.Duration
}

func (s *Settings) SetDefaults() {
	s.Upstream = gosettings.DefaultComparable(s.Upstream, "dot")
	s.ListeningAddress = gosettings.DefaultPointer(s.ListeningAddress, ":53")
	s.Block.setDefaults()
	s.Cache.setDefaults()
	s.DoH.setDefaults()
	s.DoT.setDefaults()
	s.Log.setDefaults()
	s.MiddlewareLog.setDefaults()
	s.Metrics.setDefaults()
	s.CheckDNS = gosettings.DefaultPointer(s.CheckDNS, true)
	const defaultUpdaterPeriod = 24 * time.Hour
	s.UpdatePeriod = gosettings.DefaultPointer(s.UpdatePeriod, defaultUpdaterPeriod)
}

var (
	ErrUpdatePeriodTooShort = errors.New("update period is too short")
)

func (s *Settings) Validate() (err error) {
	err = validate.IsOneOf(s.Upstream, "dot", "doh")
	if err != nil {
		return fmt.Errorf("upstream type: %w", err)
	}

	const privilegedAllowedPort = 53
	err = validate.ListeningAddress(*s.ListeningAddress, os.Getuid(), privilegedAllowedPort)
	if err != nil {
		return fmt.Errorf("listening address: %w", err)
	}

	nameToValidate := map[string]func() error{
		"block":          s.Block.validate,
		"cache":          s.Cache.validate,
		"DoH":            s.DoH.validate,
		"DoT":            s.DoT.validate,
		"log":            s.Log.validate,
		"middleware log": s.MiddlewareLog.validate,
		"metrics":        s.Metrics.validate,
		"local DNS":      s.LocalDNS.validate,
	}
	for name, validate := range nameToValidate {
		err = validate()
		if err != nil {
			return fmt.Errorf("%s settings: %w", name, err)
		}
	}

	const minUpdaterPeriod = 60 * time.Second
	if *s.UpdatePeriod != 0 && *s.UpdatePeriod < minUpdaterPeriod {
		return fmt.Errorf("%w: %s must be at least %s", ErrUpdatePeriodTooShort,
			s.UpdatePeriod, minUpdaterPeriod)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Settings:")

	node.Appendf("DNS upstream connection: %s", s.Upstream)
	node.Appendf("DNS server listening address: %s", *s.ListeningAddress)

	switch s.Upstream {
	case "dot":
		node.AppendNode(s.DoT.ToLinesNode())
	case "doh":
		node.AppendNode(s.DoH.ToLinesNode())
	default:
		panic(fmt.Sprintf("unknown upstream type: %s", s.Upstream))
	}

	node.AppendNode(s.Cache.ToLinesNode())
	node.AppendNode(s.Block.ToLinesNode())
	node.AppendNode(s.Log.ToLinesNode())
	node.AppendNode(s.MiddlewareLog.ToLinesNode())
	node.AppendNode(s.Metrics.ToLinesNode())
	node.AppendNode(s.LocalDNS.ToLinesNode())
	node.Appendf("Check DNS: %s", gosettings.BoolToYesNo(s.CheckDNS))

	if *s.UpdatePeriod == 0 {
		node.Appendf("Periodic update: disabled")
	} else {
		node.Appendf("Periodic update: every %s", *s.UpdatePeriod)
	}

	return node
}

func (s *Settings) Read(reader *reader.Reader, warner Warner) (err error) {
	warnings := checkOutdatedEnv(reader)
	for _, warning := range warnings {
		warner.Warn(warning)
	}

	s.Upstream = reader.String("UPSTREAM_TYPE")
	s.ListeningAddress = reader.Get("LISTENING_ADDRESS")

	err = s.Block.read(reader)
	if err != nil {
		return fmt.Errorf("block settings: %w", err)
	}

	err = s.Cache.read(reader)
	if err != nil {
		return fmt.Errorf("cache settings: %w", err)
	}

	err = s.DoH.read(reader)
	if err != nil {
		return fmt.Errorf("DoH settings: %w", err)
	}

	err = s.DoT.read(reader)
	if err != nil {
		return fmt.Errorf("DoT settings: %w", err)
	}

	s.Log.read(reader)

	err = s.MiddlewareLog.read(reader)
	if err != nil {
		return fmt.Errorf("middleware log settings: %w", err)
	}

	s.Metrics.read(reader)

	err = s.LocalDNS.read(reader)
	if err != nil {
		return fmt.Errorf("local DNS settings: %w", err)
	}

	s.CheckDNS, err = reader.BoolPtr("CHECK_DNS")
	if err != nil {
		return err
	}

	s.UpdatePeriod, err = reader.DurationPtr("UPDATE_PERIOD")
	if err != nil {
		return err
	}

	return nil
}
