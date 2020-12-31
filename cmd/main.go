package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/qdm12/cloudflare-dns-server/internal/dns"
	"github.com/qdm12/cloudflare-dns-server/internal/health"
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
	"github.com/qdm12/cloudflare-dns-server/internal/settings"
	"github.com/qdm12/cloudflare-dns-server/internal/splash"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/files"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/network"
)

var (
	version   string
	buildDate string
	commit    string
)

func main() {
	buildInfo := models.BuildInformation{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}
	ctx := context.Background()
	os.Exit(_main(ctx, buildInfo))
}

func _main(ctx context.Context, buildInfo models.BuildInformation) int {
	if health.IsClientMode(os.Args) {
		// Running the program in a separate instance through the Docker
		// built-in healthcheck, in an ephemeral fashion to query the
		// long running instance of the program about its status
		client := health.NewClient()
		if err := client.Query(ctx); err != nil {
			fmt.Println(err)
			return 1
		}
		return 0
	}
	fmt.Println(splash.Splash(buildInfo))

	logger, err := logging.NewLogger(logging.ConsoleEncoding, logging.InfoLevel)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paramsReader := params.NewParamsReader(logger)

	const clientTimeout = 15 * time.Second
	client := network.NewClient(clientTimeout)
	// Create configurators
	fileManager := files.NewFileManager()
	dnsConf := dns.NewConfigurator(logger, client, fileManager)

	version, err := dnsConf.Version(ctx)
	if err != nil {
		logger.Error(err)
		return 1
	}
	logger.Info("Unbound version: %s", version)

	settings, err := settings.GetSettings(paramsReader)
	if err != nil {
		logger.Error(err)
		return 1
	}
	logger.Info("Settings summary:\n" + settings.String())

	streamMerger := command.NewStreamMerger()
	go streamMerger.CollectLines(ctx,
		func(line string) { logger.Info(line) },
		func(err error) { logger.Warn(err) })

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.WithPrefix("healthcheck server: "),
		health.IsHealthy)
	wg.Add(1)
	go healthServer.Run(ctx, wg)

	waiter := command.NewWaiter()

	go unboundRunLoop(ctx, logger, dnsConf, settings, waiter, streamMerger, cancel)

	signalsCh := make(chan os.Signal, 1)
	signal.Notify(signalsCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		os.Interrupt,
	)
	select {
	case signal := <-signalsCh:
		logger.Warn("Caught OS signal %s, shutting down", signal)
		cancel()
	case <-ctx.Done():
		logger.Warn("context canceled, shutting down")
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, err := range waiter.WaitForAll(timeoutCtx) {
		logger.Warn(err)
	}
	return 1
}

func unboundRunLoop(ctx context.Context, logger logging.Logger, dnsConf dns.Configurator,
	settings models.Settings, waiter command.Waiter, streamMerger command.StreamMerger,
	fatal func(),
) {
	var timer *time.Timer
	if settings.UpdatePeriod > 0 {
		timer = time.NewTimer(settings.UpdatePeriod)
	}
	unboundCtx, unboundCancel := context.WithCancel(ctx)
	defer unboundCancel()
	for ctx.Err() == nil {
		var setupErr, startErr, waitErr error
		unboundCtx, unboundCancel, setupErr, startErr, waitErr = unboundRun(
			ctx, unboundCtx, unboundCancel, timer, dnsConf, settings, streamMerger, waiter)
		switch {
		case ctx.Err() != nil:
			logger.Warn("context canceled: exiting unbound run loop")
		case timer != nil && !timer.Stop():
			logger.Info("planned restart of unbound")
		case setupErr != nil:
			logger.Warn(setupErr)
		case startErr != nil:
			logger.Error(startErr)
			fatal()
		case waitErr != nil:
			logger.Error(waitErr)
			fatal()
		}
	}
}

func unboundRun(ctx, oldCtx context.Context, oldCancel context.CancelFunc,
	timer *time.Timer, dnsConf dns.Configurator, settings models.Settings,
	streamMerger command.StreamMerger, waiter command.Waiter) (
	newCtx context.Context, newCancel context.CancelFunc, setupErr,
	startErr, waitErr error) {
	if timer != nil {
		timer.Stop()
		timer.Reset(settings.UpdatePeriod)
	}
	if err := dnsConf.DownloadRootHints(ctx); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	if err := dnsConf.DownloadRootKey(ctx); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	if err := dnsConf.MakeUnboundConf(ctx, settings); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	newCtx, newCancel = context.WithCancel(ctx)
	oldCancel()
	stream, waitFn, err := dnsConf.Start(newCtx, settings.VerbosityDetailsLevel)
	if err != nil {
		return newCtx, newCancel, nil, err, nil
	}
	go streamMerger.Merge(newCtx, stream, command.MergeName("unbound"))
	dnsConf.UseDNSInternally(net.IP{127, 0, 0, 1}) // use Unbound
	if settings.CheckUnbound {
		if err := dnsConf.WaitForUnbound(); err != nil {
			return newCtx, newCancel, nil, err, nil
		}
	}
	waitError := make(chan error)
	waiterError := make(chan error)
	waiter.Add(func() error { //nolint:scopelint
		return <-waiterError
	})
	go func() {
		err := fmt.Errorf("unbound: %w", waitFn())
		waitError <- err
		waiterError <- err
	}()
	if timer == nil {
		waitErr := <-waitError
		return newCtx, newCancel, nil, nil, waitErr
	}
	select {
	case <-timer.C:
		return newCtx, newCancel, nil, nil, nil
	case waitErr := <-waitError:
		return newCtx, newCancel, nil, nil, waitErr
	}
}
