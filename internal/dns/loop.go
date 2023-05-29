package dns

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/setup"
	"github.com/qdm12/dns/v2/pkg/check"
	"github.com/qdm12/dns/v2/pkg/filter/mapfilter"
	"github.com/qdm12/goservices"
)

type loop struct {
	// Dependencies injected
	settings           settings.Settings
	logger             Logger
	blockBuilder       BlockBuilder
	cache              Cache
	prometheusRegistry prometheus.Registerer

	dnsServer   Service
	updateTimer *time.Timer
}

func New(settings settings.Settings, logger Logger,
	blockBuilder BlockBuilder, cache Cache,
	prometheusRegistry prometheus.Registerer) (loopService *goservices.RunWrapper, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	loop := &loop{
		settings:           settings,
		logger:             logger,
		blockBuilder:       blockBuilder,
		cache:              cache,
		prometheusRegistry: prometheusRegistry,
	}
	return goservices.NewRunWrapper("dns loop", loop.run), nil
}

func (l *loop) run(ctx context.Context, ready chan<- struct{},
	runError, stopError chan<- error) {
	err := l.runFirst(ctx)
	if err != nil {
		runError <- err
		close(runError)
		return
	}

	for {
		err := l.runSubsequent(ctx, ready)
		ready = nil // ensure we don't close it again
		switch {
		case err == nil: // planned update restart
		case errors.Is(err, errStopped):
			close(stopError)
			return
		default:
			runError <- err
			close(runError)
			return
		}
	}
}

var (
	errStopped = errors.New("stopped")
)

func (l *loop) runFirst(ctx context.Context) (err error) {
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
			return fmt.Errorf("waiting for DNS: %w", err)
		}
	}

	// Server is running or has crashed, just return nil to
	// download updated files, stop the server and start it again.
	return nil
}

func (l *loop) runSubsequent(ctx context.Context, ready chan<- struct{}) (err error) {
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

func (l *loop) setupAll(ctx context.Context, downloadBlockFiles bool) ( //nolint:ireturn
	dnsServer Service, err error) {
	filterMetrics, err := setup.FilterMetrics(l.settings.Metrics, l.prometheusRegistry)
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

	filter := mapfilter.New(filterSettings)

	server, err := setup.DNS(l.settings, l.cache,
		filter, l.logger, l.prometheusRegistry)
	if err != nil {
		return nil, fmt.Errorf("setting up DNS server: %w", err)
	}

	return server, nil
}

func (l *loop) wait(ctx context.Context, serverRunError <-chan error) (err error) {
	select {
	case <-l.updateTimer.C:
		l.logger.Info("planned periodic restart of DNS server")
		l.updateTimer.Reset(*l.settings.UpdatePeriod)
		return nil
	case <-ctx.Done():
		l.logger.Info("stopping DNS server run loop")
		err = l.dnsServer.Stop()
		if err != nil {
			return fmt.Errorf("stopping DNS server: %w", err)
		}
		return fmt.Errorf("%w", errStopped)
	case err = <-serverRunError:
		return err
	}
}
