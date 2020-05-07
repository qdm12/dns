package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/cloudflare-dns-server/internal/dns"
	"github.com/qdm12/cloudflare-dns-server/internal/healthcheck"
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
	"github.com/qdm12/cloudflare-dns-server/internal/settings"
	"github.com/qdm12/cloudflare-dns-server/internal/splash"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/files"
	libhealthcheck "github.com/qdm12/golibs/healthcheck"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/network"
)

func main() {
	if libhealthcheck.Mode(os.Args) {
		if err := healthcheck.Healthcheck(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	logger, err := logging.NewLogger(logging.ConsoleEncoding, logging.InfoLevel, -1)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paramsReader := params.NewParamsReader(logger)

	fmt.Println(splash.Splash(
		paramsReader.GetVersion(),
		paramsReader.GetVcsRef(),
		paramsReader.GetBuildDate()))

	client := network.NewClient(15 * time.Second)
	// Create configurators
	fileManager := files.NewFileManager()
	dnsConf := dns.NewConfigurator(logger, client, fileManager)

	version, err := dnsConf.Version(ctx)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	logger.Info("Unbound version: %s", version)

	settings, err := settings.GetSettings(paramsReader)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	logger.Info("Settings summary:\n" + settings.String())

	streamMerger := command.NewStreamMerger()
	go streamMerger.CollectLines(ctx,
		func(line string) { logger.Info(line) },
		func(err error) { logger.Warn(err) })

	initialDNSToUse := constants.ProviderMapping()[settings.Providers[0]]
	for _, targetIP := range initialDNSToUse.IPs {
		if settings.IPv6 && targetIP.To4() == nil {
			dnsConf.UseDNSInternally(targetIP)
			break
		} else if !settings.IPv6 && targetIP.To4() != nil {
			dnsConf.UseDNSInternally(targetIP)
			break
		}
	}

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
}

func unboundRunLoop(ctx context.Context, logger logging.Logger, dnsConf dns.Configurator,
	settings models.Settings, waiter command.Waiter, streamMerger command.StreamMerger,
	fatal func(),
) {
	timer := time.NewTimer(settings.UpdatePeriod)
	unboundCtx, unboundCancel := context.WithCancel(ctx)
	defer unboundCancel()
	for ctx.Err() == nil {
		var setupErr, startErr, waitErr error
		unboundCtx, unboundCancel, setupErr, startErr, waitErr = unboundRun(
			ctx, unboundCtx, unboundCancel, timer, dnsConf, settings, streamMerger, waiter)
		switch {
		case ctx.Err() != nil:
			logger.Warn("context canceled: exiting unbound run loop")
		case !timer.Stop():
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

func unboundRun(ctx, oldCtx context.Context, oldCancel context.CancelFunc, timer *time.Timer, dnsConf dns.Configurator, settings models.Settings,
	streamMerger command.StreamMerger, waiter command.Waiter) (newCtx context.Context, newCancel context.CancelFunc, setupErr, startErr, waitErr error) {
	timer.Stop()
	timer.Reset(settings.UpdatePeriod)
	if err := dnsConf.DownloadRootHints(); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	if err := dnsConf.DownloadRootKey(); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	if err := dnsConf.MakeUnboundConf(settings); err != nil {
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
	select {
	case <-timer.C:
		return newCtx, newCancel, nil, nil, nil
	case waitErr := <-waitError:
		return newCtx, newCancel, nil, nil, waitErr
	}
}
