package dot

import (
	"context"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/log"
	logmiddleware "github.com/qdm12/dns/pkg/middlewares/log"
	metricsmiddleware "github.com/qdm12/dns/pkg/middlewares/metrics"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Server

type Server interface {
	Run(ctx context.Context, stopped chan<- error)
}

type server struct {
	dnsServer dns.Server
	logger    log.Logger
}

func NewServer(ctx context.Context, settings ServerSettings) Server {
	settings.setDefaults()

	handler := newDNSHandler(ctx, settings)

	logMiddleware := logmiddleware.New(settings.Logger, settings.Log)
	handler = logMiddleware(handler)

	metricsMiddleware := metricsmiddleware.New(
		metricsmiddleware.Settings{Metrics: settings.Metrics})
	handler = metricsMiddleware(handler)

	return &server{
		dnsServer: dns.Server{
			Addr:    ":" + strconv.Itoa(int(settings.Port)),
			Net:     "udp",
			Handler: handler,
		},
		logger: settings.Logger,
	}
}

func (s *server) Run(ctx context.Context, stopped chan<- error) {
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
