package settings

import (
	"errors"
	"fmt"
	"os"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/gotree"
)

type MiddlewareLog struct {
	Enabled      *bool
	DirPath      string
	LogRequests  *bool
	LogResponses *bool
}

func (m *MiddlewareLog) setDefaults() {
	m.Enabled = defaults.BoolPtr(m.Enabled, false)
	m.DirPath = defaults.String(m.DirPath, "/var/log/dns/")
	m.LogRequests = defaults.BoolPtr(m.LogRequests, true)
	m.LogResponses = defaults.BoolPtr(m.LogResponses, false)
}

var ErrMiddlewareLogPathNotDirectory = errors.New("filepath specified for the middleware log is a directory")

func (m *MiddlewareLog) validate() error {
	stat, err := os.Stat(m.DirPath)
	if !os.IsNotExist(err) {
		if err != nil {
			return fmt.Errorf("directory path specified: %w", err)
		}
		if !stat.IsDir() {
			return fmt.Errorf("%w: %s", ErrMiddlewareLogPathNotDirectory, m.DirPath)
		}
	}
	return nil
}

func (m *MiddlewareLog) String() string {
	return m.ToLinesNode().String()
}

func (m *MiddlewareLog) ToLinesNode() (node *gotree.Node) {
	if !*m.Enabled {
		return gotree.New("Middleware logging: disabled")
	}

	node = gotree.New("Middleware logging:")
	node.Appendf("Log directory path: %s", m.DirPath)
	node.Appendf("Log requests: %s", boolToEnabled(*m.LogRequests))
	node.Appendf("Log responses: %s", boolToEnabled(*m.LogResponses))
	return node
}
