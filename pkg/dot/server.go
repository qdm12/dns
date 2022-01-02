package dot

import (
	"context"
	"time"

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

	handler, err := newDNSHandler(ctx, settings)
	if err != nil {
		return nil, err
	}

	logMiddleware := logmiddleware.New(settings.LogMiddleware)
	handler = logMiddleware(handler)

	metricsMiddleware := metricsmiddleware.New(
		metricsmiddleware.Settings{Metrics: settings.Metrics})
	handler = metricsMiddleware(handler)

	return &Server{
		dnsServer: dns.Server{
			Addr:    settings.Address,
			Net:     "udp",
			Handler: handler,
		},
		logger: settings.Logger,
	}, nil
}

func (s *Server) Run(ctx context.Context, stopped chan<- error) {
	go func() { // shutdown goroutine
		<-ctx.Done()

		const graceTime = 100 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), graceTime)
		defer cancel()
		if err := s.dnsServer.ShutdownContext(ctx); err != nil {
			s.logger.Error("DNS server shutdown error: " + err.Error())
		}
	}()

	s.logger.Info("DNS server listening on " + s.dnsServer.Addr)
	stopped <- s.dnsServer.ListenAndServe()
}
