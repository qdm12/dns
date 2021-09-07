package log

import (
	"reflect"
	"strings"

	"github.com/qdm12/dns/pkg/middlewares/log/format"
	"github.com/qdm12/dns/pkg/middlewares/log/format/console"
	formatnoop "github.com/qdm12/dns/pkg/middlewares/log/format/noop"
	"github.com/qdm12/dns/pkg/middlewares/log/logger"
	loggernoop "github.com/qdm12/dns/pkg/middlewares/log/logger/noop"
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

func (s *Settings) setDefaults() {
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
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	var loggerType string
	switch s.Logger.(type) { // well known types
	case *loggernoop.Logger:
		loggerType = "No-op"
	default:
		loggerType = reflect.TypeOf(s.Logger).String()
		loggerType = strings.TrimPrefix(loggerType, "*")
	}
	lines = append(lines, subSection+"Logger type: "+loggerType)

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
	lines = append(lines, subSection+"Formatter type: "+formatterType)

	return lines
}
