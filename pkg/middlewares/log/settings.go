package log

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/qdm12/dns/v2/pkg/middlewares/log/format"
	"github.com/qdm12/dns/v2/pkg/middlewares/log/logger"
	"github.com/qdm12/gotree"
)

const (
	noop    = "noop"
	console = "console"
)

type Settings struct {
	// Format is the format to serialize DNS requests
	// and responses to strings.
	// It must be the empty string if CustomFormatter is set.
	// Otherwise, it can be 'console' or 'noop',
	// and defaults to 'console' if left unset.
	// It takes precedence over CustomFormatter if set.
	Format string
	// CustomFormatter is a custom formatter to use.
	// Note Format must be empty for this custom
	// formatter to be used.
	CustomFormatter format.Formatter
	// LoggerType is the type of logger to use for the
	// logging middleware.
	// It must be the empty string if Logger is set.
	// Otherwise, it can be 'noop',
	// and defaults to 'noop' if left unset.
	// It takes precedence over Logger if set.
	LoggerType string
	// CustomLogger is a custom logger to use.
	// Note LoggerType must be empty for this custom
	// logger to be used.
	CustomLogger logger.Interface
}

func (s *Settings) SetDefaults() {
	if s.CustomFormatter == nil && s.Format == "" {
		s.Format = console
	}

	if s.CustomLogger == nil && s.LoggerType == "" {
		s.LoggerType = noop
	}
}

var (
	ErrFormatNotEmpty     = errors.New("format must be empty if custom formatter is set")
	ErrFormatNotValid     = errors.New("format is not valid")
	ErrLoggerTypeNotEmpty = errors.New("logger type must be empty if custom logger is set")
	ErrLoggerTypeNotValid = errors.New("logger type is not valid")
)

func (s *Settings) Validate() (err error) {
	if s.CustomFormatter != nil {
		if s.Format != "" {
			return ErrFormatNotEmpty
		}
	} else {
		switch s.Format {
		case noop, console:
		default:
			return fmt.Errorf("%w: %s", ErrFormatNotValid, s.Format)
		}
	}

	if s.CustomLogger != nil {
		if s.LoggerType != "" {
			return ErrLoggerTypeNotEmpty
		}
	} else if s.LoggerType != noop {
		return fmt.Errorf("%w: %s", ErrLoggerTypeNotValid, s.LoggerType)
	}

	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Log middleware settings:")

	var loggerType string
	if s.CustomLogger != nil {
		loggerType = reflect.TypeOf(s.CustomLogger).String()
		loggerType = strings.TrimPrefix(loggerType, "*")
	} else {
		loggerType = strings.Title(s.LoggerType)
	}
	node.Appendf("Logger type: %s", loggerType)

	var formatterType string
	if s.CustomFormatter != nil {
		formatterType = reflect.TypeOf(s.CustomFormatter).String()
		formatterType = strings.TrimPrefix(formatterType, "*")
	} else {
		formatterType = strings.Title(s.Format)
	}
	node.Appendf("Formatter type: %s", formatterType)

	return node
}
