package doh

import (
	"context"
	"fmt"
	"runtime"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/log"
	logmiddleware "github.com/qdm12/dns/v2/pkg/middlewares/log"
	metricsmiddleware "github.com/qdm12/dns/v2/pkg/middlewares/metrics"
)

var _ Runner = (*Server)(nil)

type Runner interface {
	Run(ctx context.Context, stopped chan<- error)
}

type Server struct {
	dnsServer dns.Server
	logger    log.Logger
}

func NewServer(ctx context.Context, settings ServerSettings) (
	server *Server, err error) {
	settings.SetDefaults()

	logger := settings.Logger

	if runtime.GOOS == "windows" {
		// TODO remove when upgrading Go to Go 1.19
		// Support was added in https://github.com/golang/go/commit/af88fb6502ceee973aaa118471c9d953a10a68e5
		logger.Warn("The Windows host cannot use the DoH server as its DNS")
	}

	var handler dns.Handler
	handler, err = newDNSHandler(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("cannot create DNS handler: %w", err)
	}

	logMiddleware := logmiddleware.New(settings.LogMiddleware)
	handler = logMiddleware(handler)

	metricsMiddleware := metricsmiddleware.New(
		metricsmiddleware.Settings{Metrics: settings.Metrics})
	handler = metricsMiddleware(handler)

	return &Server{
		dnsServer: dns.Server{
			Addr:    settings.ListeningAddress,
			Net:     "udp",
			Handler: handler,
		},
		logger: logger,
	}, nil
}

func (s *Server) Run(ctx context.Context, stopped chan<- error) {
	go func() { // shutdown goroutine
		<-ctx.Done()

		err := s.dnsServer.Shutdown()
		if err != nil {
			s.logger.Error("DNS server shutdown error: " + err.Error())
		}
	}()

	s.logger.Info("DNS server listening on " + s.dnsServer.Addr)
	stopped <- s.dnsServer.ListenAndServe()
}
