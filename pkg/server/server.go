package server

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/exchanger"
	"github.com/qdm12/goservices"
)

type Server struct {
	// Dependencies injected
	settings Settings
	logger   Logger

	// Internal state
	handlerCancel  context.CancelFunc
	state          goservices.State
	stateMutex     sync.RWMutex
	interceptStop  chan<- struct{}
	interceptDone  <-chan struct{}
	startStopMutex sync.Mutex // prevents concurrent calls to Start and Stop.
	subServers     goservices.Service
	listeningAddr  netip.AddrPort
}

func New(settings Settings) (server *Server, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Server{
		settings: settings,
		logger:   settings.Logger, // shorthand
	}, nil
}

func (s *Server) String() string {
	return "DNS server"
}

func (s *Server) Start(ctx context.Context) (runError <-chan error, startErr error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	// Lock the state in case the server is already running.
	s.stateMutex.RLock()
	state := s.state
	// no need to keep a lock on the state since the `startStopMutex`
	// prevents concurrent calls to `Start` and `Stop`.
	s.stateMutex.RUnlock()
	if state == goservices.StateRunning {
		return nil, fmt.Errorf("%s: %w", s, goservices.ErrAlreadyStarted)
	}

	s.state = goservices.StateStarting

	handlerCtx, handlerCancel := context.WithCancel(context.Background())
	defer func() {
		if startErr != nil {
			handlerCancel()
		}
	}()
	s.handlerCancel = handlerCancel

	var handler dns.Handler
	exchanger := exchanger.New(s.settings.Dialer, s.logger)
	handler = newHandler(handlerCtx, exchanger, s.logger) //nolint:contextcheck

	for _, middleware := range s.settings.Middlewares {
		handler = middleware.Wrap(handler)
	}

	udpListener, tcpListener, err := setupListeners(ctx, *s.settings.ListeningAddress)
	if err != nil {
		return nil, fmt.Errorf("setting up listeners: %w", err)
	}
	s.listeningAddr = netip.MustParseAddrPort(udpListener.LocalAddr().String())

	s.subServers, err = goservices.NewGroup(goservices.GroupSettings{
		Name: "DNS servers",
		Services: []goservices.Service{
			newSubServer(&dns.Server{
				PacketConn: udpListener,
				Handler:    handler,
			}),
			newSubServer(&dns.Server{
				Listener: tcpListener,
				Handler:  handler,
			}),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating sub servers group: %w", err)
	}

	runErrorCh := make(chan error)
	runError = runErrorCh
	subServersRunError, err := s.subServers.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting sub servers: %w", err)
	}

	interceptStop := make(chan struct{})
	s.interceptStop = interceptStop
	interceptDone := make(chan struct{})
	s.interceptDone = interceptDone
	go s.runInterceptError(interceptStop, interceptDone,
		subServersRunError, runErrorCh)

	s.logger.Info("DNS server listening on " + s.listeningAddr.String())

	s.stateMutex.Lock()
	s.state = goservices.StateRunning
	s.stateMutex.Unlock()

	return runError, nil
}

func (s *Server) runInterceptError(stop <-chan struct{}, done chan<- struct{},
	incomingError <-chan error, outgoingError chan<- error,
) {
	defer close(done)
	var err error
	select {
	case err = <-incomingError:
	case <-stop:
		return
	}
	s.stateMutex.RLock()
	state := s.state
	s.stateMutex.RUnlock()
	if state == goservices.StateStopping {
		return
	}
	s.stateMutex.Lock()
	s.state = goservices.StateCrashed
	s.stateMutex.Unlock()
	outgoingError <- err
}

func (s *Server) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.stateMutex.Lock()
	switch s.state {
	case goservices.StateRunning: // continue stopping the service
	case goservices.StateCrashed:
		s.stateMutex.Unlock()
		return fmt.Errorf("%w (crashed)", goservices.ErrAlreadyStopped)
	case goservices.StateStopped:
		s.stateMutex.Unlock()
		return fmt.Errorf("%w", goservices.ErrAlreadyStopped)
	case goservices.StateStarting, goservices.StateStopping:
		s.stateMutex.Unlock()
		panic("bad implementation code: " +
			"this code path should be unreachable for the \"" +
			fmt.Sprint(s.state) + "\" state")
	}
	s.state = goservices.StateStopping
	s.stateMutex.Unlock()

	s.handlerCancel()

	err = s.subServers.Stop()

	for _, middleware := range s.settings.Middlewares {
		middlewareErr := middleware.Stop()
		if middlewareErr != nil {
			warning := fmt.Sprintf("stopping middleware %s: %s",
				middleware, middlewareErr)
			s.logger.Warn(warning)
		}
	}

	s.state = goservices.StateStopped

	return err
}

var (
	ErrServerNotRunning      = errors.New("server not running")
	ErrListeningUDPTCPDiffer = errors.New("udp and tcp listening addresses differ")
)

func (s *Server) ListeningAddress() (address netip.AddrPort, err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.stateMutex.RLock()
	state := s.state
	s.stateMutex.RUnlock()

	if state != goservices.StateRunning {
		return netip.AddrPort{}, fmt.Errorf("%w", ErrServerNotRunning)
	}

	return s.listeningAddr, nil
}
