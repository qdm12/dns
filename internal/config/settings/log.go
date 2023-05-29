package settings

import (
	"github.com/qdm12/gosettings"
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
