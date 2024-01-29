package config

import (
	"fmt"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
	"github.com/qdm12/log"
)

type Log struct {
	Level  string
	Caller string
}

func (l Log) ToOptions() (options []log.Option) {
	logLevel, err := log.ParseLevel(l.Level)
	if err != nil {
		panic(err) // Log should be validated before
	}
	options = append(options, log.SetLevel(logLevel))
	switch l.Caller {
	case "hidden":
		options = append(options, log.SetCallerFile(false), log.SetCallerLine(false))
	case "short":
		options = append(options, log.SetCallerFile(true), log.SetCallerLine(true))
	}

	return options
}

func (l *Log) setDefaults() {
	l.Level = gosettings.DefaultComparable(l.Level, "info")
	l.Caller = gosettings.DefaultComparable(l.Caller, "hidden")
}

func (l *Log) validate() (err error) {
	_, err = log.ParseLevel(l.Level)
	if err != nil {
		return fmt.Errorf("log level: %w", err)
	}

	err = validate.IsOneOf(l.Caller, "hidden", "short")
	if err != nil {
		return fmt.Errorf("log caller: %w", err)
	}

	return nil
}

func (l *Log) String() string {
	return l.ToLinesNode().String()
}

func (l *Log) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Logging:")
	node.Appendf("Level: %s", l.Level)
	node.Appendf("Caller: %s", l.Caller)
	return node
}

func (l *Log) read(reader *reader.Reader) {
	l.Level = reader.String("LOG_LEVEL")
	l.Caller = reader.String("LOG_CALLER")
}
