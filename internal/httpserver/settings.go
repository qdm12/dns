package httpserver

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Settings struct {
	// Handler is the HTTP handler to use.
	// It must be set for settings validation to pass.
	Handler http.Handler
	// Name is the name of the server.
	// It is used for the server `String` method
	// and for logs if a logger is set.
	// It defaults to the empty string.
	Name *string
	// Address is the listening address to use.
	// It defaults to the empty string (random port) if
	// left unset.
	Address           *string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
	// Logger is the logger to use to log when the server
	// is starting and on what address it is listening.
	// It defaults to a no-op logger.
	Logger Infoer
}

func (s *Settings) SetDefaults() {
	if s.Name == nil {
		s.Name = new(string)
	}

	if s.Address == nil {
		s.Address = new(string)
	}

	if s.ReadTimeout == 0 {
		const defaultReadTimeout = 10 * time.Second
		s.ReadTimeout = defaultReadTimeout
	}

	if s.ReadHeaderTimeout == 0 {
		const defaultReadHeaderTimeout = time.Second
		s.ReadHeaderTimeout = defaultReadHeaderTimeout
	}

	if s.ShutdownTimeout == 0 {
		const defaultShutdownTimeout = 3 * time.Second
		s.ShutdownTimeout = defaultShutdownTimeout
	}

	if s.Logger == nil {
		s.Logger = new(noopLogger)
	}
}

var (
	ErrHandlerIsNil = errors.New("handler is nil")
)

func (s Settings) Validate() (err error) {
	if s.Handler == nil {
		return fmt.Errorf("%w", ErrHandlerIsNil)
	}

	if *s.Address != "" {
		_, err = net.ResolveTCPAddr("tcp", *s.Address)
		if err != nil {
			return fmt.Errorf("listening address is not valid: %w", err)
		}
	}

	return nil
}
