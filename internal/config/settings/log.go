package settings

import (
	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
	"github.com/qdm12/log"
)

type Log struct {
	Level                *log.Level
	LogRequests          *bool
	LogResponses         *bool
	LogRequestsResponses *bool
}

func (l *Log) setDefaults() {
	l.Level = defaults.LogLevelPtr(l.Level, log.LevelInfo)
	l.LogRequests = defaults.BoolPtr(l.LogRequests, false)
	l.LogResponses = defaults.BoolPtr(l.LogRequests, false)
	l.LogRequestsResponses = defaults.BoolPtr(l.LogRequests, false)
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

	var elementsLogged []string

	if *l.LogRequests {
		elementsLogged = append(elementsLogged, "requests")
	}

	if *l.LogResponses {
		elementsLogged = append(elementsLogged, "responses")
	}

	if *l.LogRequestsResponses {
		elementsLogged = append(elementsLogged, "requests and responses")
	}

	if len(elementsLogged) > 0 {
		node.Appendf("Middleware logging: %s", andStrings(elementsLogged))
	}

	return node
}
