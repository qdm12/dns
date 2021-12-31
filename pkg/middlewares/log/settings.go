package log

import (
	"reflect"
	"strings"

	"github.com/qdm12/dns/pkg/middlewares/log/format"
	"github.com/qdm12/dns/pkg/middlewares/log/format/console"
	formatnoop "github.com/qdm12/dns/pkg/middlewares/log/format/noop"
	"github.com/qdm12/dns/pkg/middlewares/log/logger"
	loggernoop "github.com/qdm12/dns/pkg/middlewares/log/logger/noop"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Formatter is the formatter to serialize DNS requests
	// and responses to strings. It defaults to the console
	// formatter if the logger is NOT a no-op logger, and to a
	// no-op formatter otherwise.
	Formatter format.Interface
	// Logger is the logger used by the DNS log middleware
	// to log requests and responses. It defaults to a No-op
	// logger implementation.
	Logger logger.Interface
}

func (s *Settings) SetDefaults() {
	if s.Formatter == nil {
		if s.Logger != nil {
			s.Formatter = console.New()
		} else {
			s.Formatter = formatnoop.New()
		}
	}

	if s.Logger == nil {
		s.Logger = loggernoop.New()
	}
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Log middleware settings:")

	var loggerType string
	switch s.Logger.(type) { // well known types
	case *loggernoop.Logger:
		loggerType = "No-op"
	default:
		loggerType = reflect.TypeOf(s.Logger).String()
		loggerType = strings.TrimPrefix(loggerType, "*")
	}
	node.Appendf("Logger type: %s", loggerType)

	var formatterType string
	switch s.Formatter.(type) { // well known types
	case *formatnoop.Formatter:
		formatterType = "No-op"
	case *console.Formatter:
		formatterType = "Console"
	default:
		formatterType = reflect.TypeOf(s.Formatter).String()
		formatterType = strings.TrimPrefix(formatterType, "*")
	}
	node.Appendf("Formatter type: %s", formatterType)

	return node
}
