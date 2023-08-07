package dot

import (
	"context"
	"fmt"
	"sync"

	"github.com/miekg/dns"
)

type Server struct {
	// Dependencies injected
	settings ServerSettings
	logger   Logger

	// Internal state
	running        bool
	runningMutex   sync.Mutex
	startStopMutex sync.Mutex // prevents concurrent calls to Start and Stop.

	// Fields set in the Start method call,
	// and shared so the Stop method can access them.
	stop      chan struct{}
	done      *sync.WaitGroup
	dnsServer dns.Server
}

func NewServer(settings ServerSettings) (server *Server, err error) {
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
	return "dns over tls server"
}

func (s *Server) Start() (runError <-chan error, startErr error) {
	s.startStopMutex.Lock()
	defer s.startStopMutex.Unlock()

	s.runningMutex.Lock()
	if s.running {
		panic("DoH server already running")
	}
	s.runningMutex.Unlock()

	handlerCtx, handlerCancel := context.WithCancel(context.Background())

	var handler dns.Handler
	var err error
	handler, err = newDNSHandler(handlerCtx, s.settings)
	if err != nil {
		handlerCancel()
		return nil, fmt.Errorf("creating DNS handler: %w", err)
	}

	for _, middleware := range s.settings.Middlewares {
		handler = middleware.Wrap(handler)
	}

	s.stop = make(chan struct{})
	s.done = new(sync.WaitGroup)
	s.dnsServer = dns.Server{
		Addr:    s.settings.ListeningAddress,
		Net:     "udp",
		Handler: handler,
	}

	var ready sync.WaitGroup
	ready.Add(1)
	s.done.Add(1)
	go func() { // cancel the handler context on a stop signal
		defer s.done.Done()
		ready.Done()
		<-s.stop
		handlerCancel()
	}()

	runErrorCh := make(chan error)
	ready.Add(1)
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		s.settings.Logger.Info("DNS server listening on " + s.dnsServer.Addr)
		ready.Done()
		err := s.dnsServer.ListenAndServe()
		s.runningMutex.Lock()
		s.running = false
		s.runningMutex.Unlock()

		select {
		case <-s.stop: // discard error
		case runErrorCh <- err:
			close(runErrorCh)
		}
	}()

	ready.Wait()

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

	close(s.stop)

	err = s.dnsServer.Shutdown()

	s.done.Wait()

	return err
}
