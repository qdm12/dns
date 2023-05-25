// Package httpserver implements an HTTP server.
package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/qdm12/dns/v2/internal/services"
)

// Server is an HTTP server implementation.
type Server struct {
	// Dependencies injected
	settings Settings

	// Internal fields
	service               services.Service
	listeningAddress      string
	listeningAddressMutex sync.RWMutex
}

// New creates a new HTTP server with a name, listening on
// the address specified and using the HTTP handler provided.
func New(settings Settings) (server *Server, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	server = &Server{
		settings: settings,
	}
	server.service = services.NewRunWrapper(*settings.Name, server.run)
	return server, nil
}

func (s *Server) String() string {
	if *s.settings.Name == "" {
		return "http server"
	}
	return *s.settings.Name + " http server"
}

// GetAddress obtains the address the HTTP server is listening on.
func (s *Server) GetAddress() (address string) {
	s.listeningAddressMutex.RLock()
	defer s.listeningAddressMutex.RUnlock()
	return s.listeningAddress
}

func (s *Server) Start() (runError <-chan error, err error) {
	return s.service.Start()
}

func (s *Server) Stop() (err error) {
	return s.service.Stop()
}

func (s *Server) run(ctx context.Context, ready chan<- struct{},
	runError, stopError chan<- error) {
	listener, err := net.Listen("tcp", *s.settings.Address)
	if err != nil {
		runError <- err
		close(runError)
		return
	}

	s.listeningAddressMutex.Lock()
	s.listeningAddress = listener.Addr().String()
	server := http.Server{
		Addr:              s.listeningAddress,
		Handler:           s.settings.Handler,
		ReadHeaderTimeout: s.settings.ReadHeaderTimeout,
		ReadTimeout:       s.settings.ReadTimeout,
	}
	s.settings.Logger.Info(fmt.Sprintf("%s listening on %s", s, s.listeningAddress))
	s.listeningAddressMutex.Unlock()

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	shutdownReady := make(chan struct{})
	shutdownDone := make(chan struct{})
	go runShutdown(shutdownCtx, shutdownReady, shutdownDone, ctx.Done(), //nolint:contextcheck
		stopError, &server, s.settings.ShutdownTimeout)
	<-shutdownReady

	close(ready)

	err = server.Serve(listener)
	if ctx.Err() != nil {
		// Server was stopped, wait for shutdown goroutine
		// to exit.
		<-shutdownDone
		return
	}

	shutdownCancel()
	<-shutdownDone

	runError <- err
	close(runError)
}

func runShutdown(ctx context.Context,
	ready, done chan<- struct{},
	shutdown <-chan struct{}, stopError chan<- error,
	server *http.Server, timeout time.Duration) {
	defer close(done)
	close(ready)
	select {
	case <-ctx.Done():
		return
	case <-shutdown:
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := server.Shutdown(shutdownCtx)
	if ctx.Err() != nil {
		return
	} else if err != nil {
		stopError <- err
	}
	close(stopError)
}
