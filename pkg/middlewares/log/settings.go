package log

import (
	"errors"
	"reflect"
	"strings"

	"github.com/qdm12/dns/v2/pkg/middlewares/log/format"
	formatnoop "github.com/qdm12/dns/v2/pkg/middlewares/log/format/noop"
	"github.com/qdm12/dns/v2/pkg/middlewares/log/logger"
	lognoop "github.com/qdm12/dns/v2/pkg/middlewares/log/logger/noop"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Formatter is a custom formatter to use.
	// It defaults to a No-op implementation.
	Formatter format.Formatter
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger logger.Interface
}

func (s *Settings) SetDefaults() {
	if s.Formatter == nil {
		s.Formatter = formatnoop.New()
	}

	if s.Logger == nil {
		s.Logger = lognoop.New()
	}
}

var (
	ErrFormatNotEmpty     = errors.New("format must be empty if custom formatter is set")
	ErrFormatNotValid     = errors.New("format is not valid")
	ErrLoggerTypeNotEmpty = errors.New("logger type must be empty if custom logger is set")
	ErrLoggerTypeNotValid = errors.New("logger type is not valid")
)

func (s *Settings) Validate() (err error) {
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Log middleware settings:")

	loggerType := reflect.TypeOf(s.Logger).String()
	loggerType = strings.TrimPrefix(loggerType, "*")
	node.Appendf("Logger type: %s", loggerType)

	formatterType := reflect.TypeOf(s.Formatter).String()
	formatterType = strings.TrimPrefix(formatterType, "*")
	node.Appendf("Formatter type: %s", formatterType)

	return node
}
