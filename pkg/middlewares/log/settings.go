package log

import (
	"errors"
	"reflect"
	"strings"

	"github.com/qdm12/dns/v2/pkg/middlewares/log/logger/noop"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type Settings struct {
	// Logger is the logger to use.
	// It defaults to a No-op implementation.
	Logger Logger
}

func (s *Settings) SetDefaults() {
	s.Logger = gosettings.DefaultInterface(s.Logger, noop.New())
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

	return node
}
