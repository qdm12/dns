package dns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/setup"
	"github.com/qdm12/dns/v2/pkg/check"
	"github.com/qdm12/dns/v2/pkg/middlewares/filter/mapfilter"
)

type Loop struct {
	// Dependencies injected
	settings           settings.Settings
	logger             Logger
	blockBuilder       BlockBuilder
	cache              Cache
	prometheusRegistry PrometheusRegistry

	dnsServer   Service
	updateTimer *time.Timer
	runCancel   context.CancelFunc
	runDone     chan struct{}
}

func New(settings settings.Settings, logger Logger,
	blockBuilder BlockBuilder, cache Cache,
	prometheusRegistry PrometheusRegistry) (loop *Loop, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	return &Loop{
		settings:           settings,
		logger:             logger,
		blockBuilder:       blockBuilder,
		cache:              cache,
		prometheusRegistry: prometheusRegistry,
	}, nil
}

func (l *Loop) String() string {
	return "dns loop"
}

func (l *Loop) Start(ctx context.Context) (
	runError <-chan error, err error) {
	err = l.runFirst(ctx)
	if err != nil {
		return nil, err
	}

	runErrorBidirectional := make(chan error)
	runError = runErrorBidirectional

	var runCtx context.Context
	runCtx, l.runCancel = context.WithCancel(context.Background())
	l.runDone = make(chan struct{})
	ready := make(chan struct{})

	go func() {
		defer close(l.runDone)
		for runCtx.Err() == nil {
			err := l.runSubsequent(runCtx, ready)
			switch {
			case err == nil: // planned update restart
			case errors.Is(err, runCtx.Err()):
				return
			default:
				runErrorBidirectional <- err
				close(runErrorBidirectional)
				return
			}
		}
	}()
	<-ready
	ready = nil

	return runError, nil
}

func (l *Loop) Stop() (err error) {
	l.runCancel()
	<-l.runDone
	return l.dnsServer.Stop()
}

func (l *Loop) runFirst(ctx context.Context) (err error) {
	const downloadBlockFiles = false
	l.dnsServer, err = l.setupAll(ctx, downloadBlockFiles)
	if err != nil {
		return fmt.Errorf("setting up DNS server: %w", err)
	}

	l.logger.Info("starting DNS server")
	_, err = l.dnsServer.Start()
	if err != nil {
		return fmt.Errorf("starting dns server: %w", err)
	}

	if *l.settings.CheckDNS {
		err = check.WaitForDNS(ctx, check.Settings{})
		if err != nil {
			_ = l.dnsServer.Stop()
			return fmt.Errorf("waiting for DNS: %w", err)
		}
	}

	// Server is running or has crashed, just return nil to
	// download updated files, stop the server and start it again.
	return nil
}

func (l *Loop) runSubsequent(ctx context.Context,
	ready chan<- struct{}) (err error) {
	const downloadBlockFiles = true
	newDNSServer, err := l.setupAll(ctx, downloadBlockFiles)
	if err != nil {
		return fmt.Errorf("setting up DNS server: %w", err)
	}

	err = l.dnsServer.Stop()
	if err != nil {
		return fmt.Errorf("stopping DNS server: %w", err)
	}
	l.dnsServer = newDNSServer

	l.logger.Info("starting DNS server")
	serverRunError, startErr := l.dnsServer.Start()
	if startErr != nil {
		return fmt.Errorf("starting dns server: %w", startErr)
	}

	if *l.settings.CheckDNS {
		err = check.WaitForDNS(ctx, check.Settings{})
		if err != nil {
			_ = l.dnsServer.Stop()
			return fmt.Errorf("waiting for DNS: %w", err)
		}
	}

	if ready != nil {
		close(ready)
	}

	l.updateTimer = time.NewTimer(time.Hour)
	l.updateTimer.Stop()
	if *l.settings.UpdatePeriod > 0 {
		l.updateTimer.Reset(*l.settings.UpdatePeriod)
	}

	return l.wait(ctx, serverRunError)
}

func (l *Loop) setupAll(ctx context.Context, downloadBlockFiles bool) ( //nolint:ireturn
	dnsServer Service, err error) {
	filterMetrics, err := setup.BuildFilterMetrics(l.settings.Metrics, l.prometheusRegistry)
	if err != nil {
		return nil, fmt.Errorf("setting up filter metrics: %w", err)
	}

	filterSettings := mapfilter.Settings{
		Metrics: filterMetrics,
	}

	if downloadBlockFiles {
		l.logger.Info("downloading and building DNS block lists")
		result := l.blockBuilder.BuildAll(ctx)
		for _, err := range result.Errors {
			l.logger.Warn(err.Error())
		}
		l.logger.Info(fmt.Sprint(len(result.BlockedHostnames)) + " hostnames blocked overall")
		l.logger.Info(fmt.Sprint(len(result.BlockedIPs)) + " IP addresses blocked overall")
		l.logger.Info(fmt.Sprint(len(result.BlockedIPPrefixes)) + " IP networks blocked overall")
		filterSettings.Update.IPs = result.BlockedIPs
		filterSettings.Update.IPPrefixes = result.BlockedIPPrefixes
		filterSettings.Update.BlockHostnames(result.BlockedHostnames)
	}

	filter, err := mapfilter.New(filterSettings)
	if err != nil {
		return nil, fmt.Errorf("setting up filter: %w", err)
	}

	server, err := setup.DNS(l.settings, l.cache,
		filter, l.logger, l.prometheusRegistry)
	if err != nil {
		return nil, fmt.Errorf("setting up DNS server: %w", err)
	}

	return server, nil
}

func (l *Loop) wait(ctx context.Context, serverRunError <-chan error) (err error) {
	select {
	case <-l.updateTimer.C:
		l.logger.Info("planned periodic restart of DNS server")
		l.updateTimer.Reset(*l.settings.UpdatePeriod)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case err = <-serverRunError:
		return err
	}
}
