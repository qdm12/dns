package dot

import (
	"context"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/golibs/logging"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Server

type Server interface {
	Run(ctx context.Context, stopped chan<- error)
}

type server struct {
	dnsServer dns.Server
	logger    logging.Logger
}

func NewServer(ctx context.Context, logger logging.Logger,
	settings ServerSettings) Server {
	settings.setDefaults()

	handler := newDNSHandler(ctx, logger, settings)
	logMiddleware := log.New(logger, settings.Log)
	handler = logMiddleware(handler)

	return &server{
		dnsServer: dns.Server{
			Addr:    ":" + strconv.Itoa(int(settings.Port)),
			Net:     "udp",
			Handler: handler,
		},
		logger: logger,
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
