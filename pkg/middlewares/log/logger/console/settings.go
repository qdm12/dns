package console

import (
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/qdm12/gotree"
)

type Settings struct {
	// Writer is the writer to write logs to.
	// It defaults to os.Stdout.
	Writer io.Writer
	// LogRequests indicates requests should be logged.
	// If LogResponses is enabled as well, each request
	// is logged together with its matching response.
	LogRequests *bool
	// LogResponses indicates responses should be logged.
	// If LogRequests is enabled as well, each request
	// is logged together with its matching response.
	LogResponses *bool
}

func (s *Settings) SetDefaults() {
	if s.Writer == nil {
		s.Writer = os.Stdout
	}

	if s.LogRequests == nil {
		s.LogRequests = boolPtr(true)
	}

	if s.LogResponses == nil {
		s.LogResponses = boolPtr(false)
	}
}

func boolPtr(b bool) *bool { return &b }

func (s *Settings) Validate() (err error) {
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Middleware logger settings:")

	writerType := reflect.TypeOf(s.Writer).String()
	writerType = strings.TrimPrefix(writerType, "*")
	node.Appendf("Writer type: %s", writerType)

	node.Appendf("Log requests: %t", *s.LogRequests)
	node.Appendf("Log responses: %t", *s.LogResponses)

	return node
}
