package settings

import (
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
)

// Settings contain settings to configure the entire program
// and to live patch the program from external sources.
type Settings struct {
	// Upstream is the DNS upstream connection type
	// and can be either 'dot' or 'doh'.
	// It defaults to 'dot' if left uset.
	Upstream         string
	ListeningAddress string
	Block            Block
	Cache            Cache
	DoH              DoH
	DoT              DoT
	Log              Log
	MiddlewareLog    MiddlewareLog
	Metrics          Metrics
	CheckDNS         *bool
	UpdatePeriod     *time.Duration
}

func (s *Settings) SetDefaults() {
	s.Upstream = defaults.String(s.Upstream, "dot")
	s.ListeningAddress = defaults.String(s.ListeningAddress, ":53")
	s.Block.setDefaults()
	s.Cache.setDefaults()
	s.DoH.setDefaults()
	s.DoT.setDefaults()
	s.Log.setDefaults()
	s.MiddlewareLog.setDefaults()
	s.Metrics.setDefaults()
	s.CheckDNS = defaults.BoolPtr(s.CheckDNS, true)
	const defaultUpdaterPeriod = 24 * time.Hour
	s.UpdatePeriod = defaults.DurationPtr(s.UpdatePeriod, defaultUpdaterPeriod)
}

var (
	ErrUpdatePeriodTooShort = errors.New("update period must be at least one minute")
)

func (s *Settings) Validate() (err error) {
	err = checkIsOneOf(s.Upstream, "dot", "doh")
	if err != nil {
		return fmt.Errorf("upstream type is unknown: %w", err)
	}

	err = checkListeningAddress(s.ListeningAddress)
	if err != nil {
		return fmt.Errorf("listening address is not valid: %w", err)
	}

	nameToValidate := map[string]func() error{
		"block":          s.Block.validate,
		"cache":          s.Cache.validate,
		"DoH":            s.DoH.validate,
		"DoT":            s.DoT.validate,
		"log":            s.Log.validate,
		"middleware log": s.MiddlewareLog.validate,
		"metrics":        s.Metrics.validate,
	}
	for name, validate := range nameToValidate {
		err = validate()
		if err != nil {
			return fmt.Errorf("%s settings: %w", name, err)
		}
	}

	const minUpdaterPeriod = 60 * time.Second
	if *s.UpdatePeriod != 0 && *s.UpdatePeriod < minUpdaterPeriod {
		return fmt.Errorf("%w: %s", ErrUpdatePeriodTooShort, s.UpdatePeriod)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Settings:")

	node.Appendf("DNS upstream connection: %s", s.Upstream)
	node.Appendf("DNS server listening address: %s", s.ListeningAddress)

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
	node.Appendf("Check DNS: %s", boolToEnabled(*s.CheckDNS))

	if *s.UpdatePeriod == 0 {
		node.Appendf("Periodic update: disabled")
	} else {
		node.Appendf("Periodic update: every %s", *s.UpdatePeriod)
	}

	return node
}
