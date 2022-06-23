// Package prometheus offers a Setup function to setup a Prometheus
// HTTP server together with all the metrics registered.
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qdm12/dns/v2/internal/config/settings"
)

type Logger interface {
	Info(s string)
	Warn(s string)
	Error(s string)
}

func Setup(settings settings.Prometheus, gatherer prometheus.Gatherer,
	logger Logger) (server *Server) {
	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		// ErrorLog: logger, // TODO
	})
	server = &Server{
		address: settings.ListeningAddress,
		handler: handler,
		logger:  logger,
	}

	return server
}
