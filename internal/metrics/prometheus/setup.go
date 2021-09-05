// Package prometheus offers a Setup function to setup a Prometheus
// HTTP server together with all the metrics registered.
package prometheus

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qdm12/dns/internal/config"
	cache "github.com/qdm12/dns/pkg/cache/metrics/prometheus"
	doh "github.com/qdm12/dns/pkg/doh/metrics/prometheus"
	dot "github.com/qdm12/dns/pkg/dot/metrics/prometheus"
	filter "github.com/qdm12/dns/pkg/filter/metrics/prometheus"
	promshared "github.com/qdm12/dns/pkg/metrics/prometheus"
	middleware "github.com/qdm12/dns/pkg/middlewares/metrics/prometheus"
)

var (
	ErrCache      = errors.New("cannot setup cache metrics")
	ErrMiddleware = errors.New("cannot setup middleware metrics")
	ErrDOT        = errors.New("cannot setup DoT metrics")
	ErrDOH        = errors.New("cannot setup DoH metrics")
)

type Logger interface {
	Info(s string)
	Warn(s string)
	Error(s string)
}

func Setup(settings config.Prometheus, logger Logger) (
	server *Server,
	cacheMetrics *cache.Metrics,
	filterMetrics *filter.Metrics,
	dotMetrics *dot.Metrics,
	dohMetrics *doh.Metrics,
	err error) {
	promRegistry := prometheus.NewRegistry()

	metricsSettings := promshared.Settings{
		Prefix:   settings.Subsystem,
		Registry: promRegistry,
	}

	cacheMetrics, err = cache.New(
		cache.Settings{Prometheus: metricsSettings})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: %s", ErrCache, err)
	}

	filterMetrics, err = filter.New(
		filter.Settings{Prometheus: metricsSettings})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: %s", ErrCache, err)
	}

	middlewareMetrics, err := middleware.New(
		middleware.Settings{Prometheus: metricsSettings})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: %s", ErrMiddleware, err)
	}

	dotMetrics, err = dot.New(dot.Settings{
		Prometheus:        metricsSettings,
		MiddlewareMetrics: middlewareMetrics,
	})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: %s", ErrDOT, err)
	}

	dohMetrics, err = doh.New(doh.Settings{
		Prometheus:        metricsSettings,
		DoTDialMetrics:    dotMetrics,
		MiddlewareMetrics: middlewareMetrics,
	})
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("%w: %s", ErrDOH, err)
	}

	handler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{})
	server = &Server{
		address: settings.Address,
		handler: handler,
		logger:  logger,
	}

	return server, cacheMetrics, filterMetrics, dotMetrics, dohMetrics, nil
}
