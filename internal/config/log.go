package config

import (
	"fmt"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
	"github.com/qdm12/log"
)

type Log struct {
	Level *log.Level
}

func (l *Log) setDefaults() {
	l.Level = gosettings.DefaultPointer(l.Level, log.LevelInfo)
}

func (l *Log) validate() error {
	return nil
}

func (l *Log) String() string {
	return l.ToLinesNode().String()
}

func (l *Log) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Logging:")
	node.Appendf("Level: %s", l.Level)
	return node
}

func (l *Log) read(reader *reader.Reader) (err error) {
	levelString := reader.String("LOG_LEVEL")
	if levelString == "" {
		return nil
	}

	levelValue, err := log.ParseLevel(levelString)
	if err != nil {
		return fmt.Errorf("parsing log level: %w", err)
	}
	l.Level = &levelValue

	return nil
}
