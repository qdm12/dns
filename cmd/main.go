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
	os.Exit(_main(ctx, buildInfo, os.Args))
}

func _main(ctx context.Context, buildInfo models.BuildInformation, args []string) int {
	if health.IsClientMode(args) {
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	paramsReader := params.NewParamsReader(logger)

	const clientTimeout = 15 * time.Second
	client := network.NewClient(clientTimeout)
	// Create configurators
	fileManager := files.NewFileManager()
	dnsConf := dns.NewConfigurator(logger, fileManager)

	if len(args) > 1 && args[1] == "build" {
		if err := dnsConf.DownloadRootHints(ctx, client); err != nil {
			logger.Error(err)
			return 1
		}
		if err := dnsConf.DownloadRootKey(ctx, client); err != nil {
			logger.Error(err)
			return 1
		}
		return 0
	}

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

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.WithPrefix("healthcheck server: "),
		health.IsHealthy)
	wg.Add(1)
	go healthServer.Run(ctx, wg)

	dnsConf.UseDNSInternally(net.IP{127, 0, 0, 1}) // use Unbound
	wg.Add(1)
	go unboundRunLoop(ctx, wg, settings, logger, dnsConf, streamMerger, client, cancel)

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

	waited := make(chan struct{})
	timer := time.NewTimer(time.Second)
	go func() {
		wg.Wait()
		close(waited)
	}()
	select {
	case <-waited:
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
		logger.Error("shutdown timed out, force quitting")
	}

	return 1
}

func unboundRunLoop(ctx context.Context, wg *sync.WaitGroup, settings models.Settings,
	logger logging.Logger, dnsConf dns.Configurator, streamMerger command.StreamMerger,
	client network.Client, fatal func(),
) {
	defer wg.Done()
	defer logger.Info("unbound loop exited")
	timer := time.NewTimer(time.Hour)

	unboundCtx, unboundCancel := context.WithCancel(ctx)
	defer unboundCancel()

	firstRun := true

	for ctx.Err() == nil {
		timer.Stop()
		if settings.UpdatePeriod > 0 {
			timer.Reset(settings.UpdatePeriod)
		}

		var setupErr, startErr, waitErr error
		unboundCtx, unboundCancel, setupErr, startErr, waitErr = unboundRun(
			ctx, unboundCtx, unboundCancel, timer, dnsConf, settings, streamMerger,
			client, firstRun)
		switch {
		case ctx.Err() != nil:
			logger.Warn("context canceled: exiting unbound run loop")
		case !timer.Stop():
			logger.Info("planned restart of unbound")
		case setupErr != nil:
			logAndWait(ctx, logger, setupErr)
		case firstRun:
			logger.Info("restarting Unbound the first time to get updated files")
			firstRun = false
		case startErr != nil:
			logger.Error(startErr)
			fatal()
		case waitErr != nil:
			logger.Error(waitErr)
			fatal()
		}
	}
}

func logAndWait(ctx context.Context, logger logging.Logger, err error) {
	const wait = 10 * time.Second
	logger.Error("%s, retrying in %s", err, wait)
	timer := time.NewTimer(wait)
	select {
	case <-timer.C:
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	}
}

func unboundRun(ctx, oldCtx context.Context, oldCancel context.CancelFunc,
	timer *time.Timer, dnsConf dns.Configurator, settings models.Settings,
	streamMerger command.StreamMerger,
	client network.Client, firstRun bool) (
	newCtx context.Context, newCancel context.CancelFunc, setupErr,
	startErr, waitErr error) {
	var hostnamesLines, ipsLines []string
	if !firstRun {
		if err := dnsConf.DownloadRootHints(ctx, client); err != nil {
			return oldCtx, oldCancel, err, nil, nil
		}
		if err := dnsConf.DownloadRootKey(ctx, client); err != nil {
			return oldCtx, oldCancel, err, nil, nil
		}
		blockedIPs := append(settings.BlockedIPs, settings.PrivateAddresses...)
		hostnamesLines, ipsLines = dnsConf.BuildBlocked(ctx, client,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.BlockedHostnames, blockedIPs, settings.AllowedHostnames,
		)
	}
	if err := dnsConf.MakeUnboundConf(settings, hostnamesLines, ipsLines); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	newCtx, newCancel = context.WithCancel(ctx)
	oldCancel()
	stream, waitFn, err := dnsConf.Start(newCtx, settings.VerbosityDetailsLevel)
	if err != nil {
		return newCtx, newCancel, nil, err, nil
	}
	go streamMerger.Merge(newCtx, stream, command.MergeName("unbound"))
	if settings.CheckUnbound {
		if err := dnsConf.WaitForUnbound(ctx); err != nil {
			return newCtx, newCancel, nil, err, nil
		}
	}

	waitError := make(chan error)
	go func() {
		err := waitFn() // blocking
		waitError <- err
	}()

	if firstRun { // force restart
		return newCtx, newCancel, nil, nil, nil
	}

	select {
	case <-timer.C:
		return newCtx, newCancel, nil, nil, nil
	case waitErr := <-waitError:
		close(waitError)
		return newCtx, newCancel, nil, nil, waitErr
	}
}
