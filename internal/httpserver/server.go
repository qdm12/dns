// Package httpserver implements an HTTP server.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// Server is an HTTP server implementation.
type Server struct {
	// Dependencies injected
	settings Settings

	// Internal state
	running        bool
	runningMutex   sync.Mutex
	startStopMutex sync.Mutex

	// Fields set in the Start method call,
	// and shared so the Stop method can access them.
	stop   chan struct{}
	done   chan struct{}
	server http.Server
}

// New creates a new HTTP server with a name, listening on
// the address specified and using the HTTP handler provided.
func New(settings Settings) (server *Server, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Server{
		settings: settings,
	}, nil
}

func (s *Server) String() string {
	if *s.settings.Name == "" {
		return "http server"
	}
	return *s.settings.Name + " http server"
}

// Start starts the HTTP server and returns a read only error channel
// and an eventual start error.
// If an error is encountered during the server run, this one is sent
// in the `runError` channel, and the channel is then closed.
// A caller should not call the `Stop` method after an error has been read
// from the `runError` channel, since the server is already stopped.
// If this method is called but the server is already started, the function panics.
func (s *Server) Start() (runError <-chan error, err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.runningMutex.Lock()
	if s.running {
		panic("server already started")
	}
	s.runningMutex.Unlock()

	runErrorCh := make(chan error)

	listener, err := net.Listen("tcp", *s.settings.Address)
	if err != nil {
		return nil, err
	}

	s.server = http.Server{
		Addr:              listener.Addr().String(),
		Handler:           s.settings.Handler,
		ReadHeaderTimeout: s.settings.ReadHeaderTimeout,
		ReadTimeout:       s.settings.ReadTimeout,
	}
	s.settings.Logger.Info(fmt.Sprintf("%s listening on %s", s, s.server.Addr))

	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	ready := make(chan struct{})
	go func(ready, done chan<- struct{}, listener net.Listener, runErrorCh chan<- error) {
		defer close(done)
		close(ready)

		err := s.server.Serve(listener)

		// Set the running state to false as soon as the server
		// returns an error.
		s.runningMutex.Lock()
		s.running = false
		s.runningMutex.Unlock()

		// If the server was stopped, do not write the error to the run error channel.
		// Otherwise, write the error to the run error channel.
		// The reason for this select is to address races where the server would crash
		// at the same time it is stopped, and the caller stops listening for an error
		// on the run error channel. In that case we don't want to let this goroutine
		// be stuck and create a memory leak.
		select {
		case <-s.stop: // discard error
			return
		default:
			runErrorCh <- err
			close(runErrorCh)
		}
	}(ready, s.done, listener, runErrorCh)
	<-ready

	s.runningMutex.Lock()
	s.running = true
	s.runningMutex.Unlock()

	return runErrorCh, nil
}

var (
	ErrServerNotRunning = errors.New("server is not running")
)

// GetAddress obtains the address the HTTP server is listening on.
func (s *Server) GetAddress() (address string, err error) {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.running {
		return "", fmt.Errorf("%w", ErrServerNotRunning)
	}

	return s.server.Addr, nil
}

// Stop stops the server within the given shutdown timeout.
func (s *Server) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), s.settings.ShutdownTimeout)
	defer cancel()

	s.runningMutex.Lock()
	running := s.running //nolint:ifshort
	s.runningMutex.Unlock()
	if !running { // server crashed whilst we were stopping it
		return nil
	}

	close(s.stop)

	err = s.server.Shutdown(shutdownCtx)
	<-s.done
	return err
}
