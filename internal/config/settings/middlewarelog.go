package settings

import (
	"errors"
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type MiddlewareLog struct {
	Enabled      *bool
	DirPath      string
	LogRequests  *bool
	LogResponses *bool
}

func (m *MiddlewareLog) setDefaults() {
	m.Enabled = gosettings.DefaultPointer(m.Enabled, false)
	m.DirPath = gosettings.DefaultString(m.DirPath, "/var/log/dns/")
	m.LogRequests = gosettings.DefaultPointer(m.LogRequests, true)
	m.LogResponses = gosettings.DefaultPointer(m.LogResponses, false)
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
	node.Appendf("Log requests: %s", gosettings.BoolToYesNo(m.LogRequests))
	node.Appendf("Log responses: %s", gosettings.BoolToYesNo(m.LogResponses))
	return node
}
