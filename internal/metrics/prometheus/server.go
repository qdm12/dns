package prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/goservices/httpserver"
)

func New(settings config.Prometheus, gatherer Gatherer,
	logger Logger) (server *httpserver.Server, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	handlerOptions := promhttp.HandlerOpts{
		ErrorLog: &promLogger{logger: logger},
	}
	httpSettings := httpserver.Settings{
		Name:    stringPtr("prometheus"),
		Handler: promhttp.HandlerFor(gatherer, handlerOptions),
		Address: stringPtr(settings.ListeningAddress),
		Logger:  logger,
	}
	return httpserver.New(httpSettings)
}

func stringPtr(s string) *string { return &s }
