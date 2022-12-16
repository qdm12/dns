package dns

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/setup"
	"github.com/qdm12/dns/v2/pkg/check"
	"github.com/qdm12/dns/v2/pkg/filter/mapfilter"
)

type Loop struct {
	// Dependencies injected
	settings           settings.Settings
	logger             Logger
	blockBuilder       BlockBuilder
	cache              Cache
	prometheusRegistry prometheus.Registerer

	// Internal state
	running      bool
	runningMutex sync.Mutex
	// mutex prevents concurrent calls to Start and Stop.
	mutex sync.Mutex

	// Fields set in the Start method call,
	// and shared so the Stop method can access them.
	stop      chan struct{}
	done      chan struct{}
	dnsServer Service
}

func New(settings settings.Settings, logger Logger,
	blockBuilder BlockBuilder, cache Cache,
	prometheusRegistry prometheus.Registerer) (loop *Loop, err error) {
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

func (l *Loop) Start() (runError <-chan error, startErr error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.runningMutex.Lock()
	if l.running {
		panic("dns loop already started")
	}
	l.runningMutex.Unlock()

	stopStopToCtx := make(chan struct{})
	ctx, stopToCtxDone := forwardStopAsCancel(stopStopToCtx)
	defer func() {
		close(stopStopToCtx)
		<-stopToCtxDone
	}()

	var err error
	const downloadBlockFiles = false
	l.dnsServer, err = l.setupAll(ctx, downloadBlockFiles)
	if err != nil {
		return nil, fmt.Errorf("setting up DNS server: %w", err)
	}

	l.logger.Info("starting DNS server")
	serverRunError, err := l.dnsServer.Start()
	if err != nil {
		return nil, fmt.Errorf("starting dns server: %w", err)
	}

	if *l.settings.CheckDNS {
		err = check.WaitForDNS(ctx, check.Settings{})
		if err != nil {
			return nil, fmt.Errorf("waiting for DNS: %w", err)
		}
	}

	// Make sure it didn't crash already.
	select {
	case err = <-serverRunError:
		return nil, fmt.Errorf("running DNS server: %w", err)
	default:
	}

	runErrorCh := make(chan error)

	ready := make(chan struct{})
	l.stop = make(chan struct{})
	l.done = make(chan struct{})
	go l.run(ready, runErrorCh)
	<-ready

	l.runningMutex.Lock()
	l.running = true
	l.runningMutex.Unlock()

	return runErrorCh, nil
}

func (l *Loop) Stop() (err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.runningMutex.Lock()
	if !l.running {
		l.runningMutex.Unlock()
		return nil
	}
	l.runningMutex.Unlock()

	close(l.stop)
	<-l.done
	return nil
}

func (l *Loop) run(ready chan<- struct{}, runError chan<- error) {
	defer close(l.done)

	ctx, stopToCtxDone := forwardStopAsCancel(l.stop)
	defer func() {
		<-stopToCtxDone
	}()

	updateTimer := time.NewTimer(time.Hour)
	updateTimer.Stop()
	updatePeriod := *l.settings.UpdatePeriod
	if updatePeriod > 0 {
		updateTimer.Reset(updatePeriod)
	}

	// DNS server is already up from the `Start` method call preceding
	// this function call.
	close(ready)

	l.logger.Info("restarting DNS server the first time to use updated files")

	downloadBlockFiles := true
	for {
		err := l.runOnce(ctx, downloadBlockFiles, updateTimer)
		if err == nil {
			// only download block files on the second DNS run (first run here)
			downloadBlockFiles = false
			continue
		}

		l.runningMutex.Lock()
		l.running = false
		l.runningMutex.Unlock()

		if errors.Is(err, errStopped) {
			break
		}

		select {
		case <-l.stop: // discard error
			return
		default:
			runError <- err
			close(runError)
		}
		break
	}
}

var (
	errStopped = errors.New("stopped")
)

func (l *Loop) runOnce(ctx context.Context, downloadBlockFiles bool,
	updateTimer *time.Timer) (err error) {
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

	select {
	case <-updateTimer.C:
		l.logger.Info("planned periodic restart of DNS server")
		updateTimer.Reset(*l.settings.UpdatePeriod)
		return nil
	case <-l.stop:
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

func (l *Loop) setupAll(ctx context.Context, downloadBlockFiles bool) ( //nolint:ireturn
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

func forwardStopAsCancel(stop <-chan struct{}) (ctx context.Context, done <-chan struct{}) {
	stopToCtxReady := make(chan struct{})
	stopToCtxDone := make(chan struct{})
	ctx, ctxCancel := context.WithCancel(context.Background())
	go func(cancel context.CancelFunc, ready, done chan<- struct{}) {
		defer close(done)
		defer cancel()
		close(ready)
		<-stop
	}(ctxCancel, stopToCtxReady, stopToCtxDone)
	<-stopToCtxReady
	return ctx, stopToCtxDone
}
