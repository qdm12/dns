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
	running        bool
	runningMutex   sync.Mutex
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

	s.runningMutex.Lock()
	if s.running {
		panic("server already running")
	}
	s.runningMutex.Unlock()

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

	runErrorCh, err := s.subServers.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting sub servers: %w", err)
	}

	s.logger.Info("DNS server listening on " + s.listeningAddr.String())

	s.runningMutex.Lock()
	s.running = true
	s.runningMutex.Unlock()

	return runErrorCh, nil
}

func (s *Server) Stop() (err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.runningMutex.Lock()
	running := s.running //nolint:ifshort
	s.runningMutex.Unlock()
	if !running { // server crashed whilst we were stopping it
		return nil
	}

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

	return err
}

var (
	ErrServerNotRunning      = errors.New("server not running")
	ErrListeningUDPTCPDiffer = errors.New("udp and tcp listening addresses differ")
)

func (s *Server) ListeningAddress() (address netip.AddrPort, err error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	if !s.running {
		return netip.AddrPort{}, fmt.Errorf("%w", ErrServerNotRunning)
	}

	return s.listeningAddr, nil
}
