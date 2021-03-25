package doh

import (
	"context"
	"runtime"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/golibs/logging"
)

type Server interface {
	Run(ctx context.Context, stopped chan<- struct{})
}

type server struct {
	dnsServer dns.Server
	logger    logging.Logger
}

func NewServer(ctx context.Context, logger logging.Logger,
	options ...Option) Server {
	if runtime.GOOS == "windows" {
		panic("DoH server cannot work on Windows")
	}
	return &server{
		dnsServer: dns.Server{
			Addr:    ":53",
			Net:     "udp",
			Handler: newDNSHandler(ctx, logger, options...),
		},
		logger: logger,
	}
}

func (s *server) Run(ctx context.Context, stopped chan<- struct{}) {
	defer close(stopped)

	go func() { // shutdown goroutine
		<-ctx.Done()

		const graceTime = 100 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), graceTime)
		defer cancel()
		if err := s.dnsServer.ShutdownContext(ctx); err != nil {
			s.logger.Error("DNS server shutdown error: ", err)
		}
	}()

	s.logger.Info("DNS server listening on :53")
	if err := s.dnsServer.ListenAndServe(); err != nil {
		s.logger.Error("DNS server crashed: ", err)
	}
	s.logger.Warn("DNS server stopped")
}
