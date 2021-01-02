package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	nativeos "os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/qdm12/dns/internal/health"
	"github.com/qdm12/dns/internal/models"
	"github.com/qdm12/dns/internal/params"
	"github.com/qdm12/dns/internal/settings"
	"github.com/qdm12/dns/internal/splash"
	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/os"
	"github.com/qdm12/updated/pkg/dnscrypto"
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
	os := os.New()
	nativeos.Exit(_main(ctx, buildInfo, nativeos.Args, os))
}

func _main(ctx context.Context, buildInfo models.BuildInformation, args []string, os os.OS) int {
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
	client := &http.Client{Timeout: clientTimeout}
	// Create configurators
	dnsCrypto := dnscrypto.New(client, "", "") // TODO checksums for build
	const unboundEtcDir = "/unbound"
	const unboundPath = "/unbound/unbound"
	const cacertsPath = "/unbound/ca-certificates.crt"
	dnsConf := unbound.NewConfigurator(logger, os.OpenFile, dnsCrypto, unboundEtcDir, unboundPath, cacertsPath)

	if len(args) > 1 && args[1] == "build" {
		if err := dnsConf.SetupFiles(ctx); err != nil {
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

	localIP := net.IP{127, 0, 0, 1}
	logger.Info("using DNS address %s internally", localIP.String())
	dnsConf.UseDNSInternally(localIP) // use Unbound
	wg.Add(1)
	go unboundRunLoop(ctx, wg, settings, logger, dnsConf, streamMerger, client, cancel)

	signalsCh := make(chan nativeos.Signal, 1)
	signal.Notify(signalsCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		nativeos.Interrupt,
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
	logger logging.Logger, dnsConf unbound.Configurator, streamMerger command.StreamMerger,
	client *http.Client, fatal func(),
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
			logger, client, firstRun)
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
	timer *time.Timer, dnsConf unbound.Configurator, settings models.Settings,
	streamMerger command.StreamMerger, logger logging.Logger,
	client *http.Client, firstRun bool) (
	newCtx context.Context, newCancel context.CancelFunc, setupErr,
	startErr, waitErr error) {
	var hostnamesLines, ipsLines []string
	if !firstRun {
		logger.Info("downloading DNSSEC root hints and named root")
		if err := dnsConf.SetupFiles(ctx); err != nil {
			return oldCtx, oldCancel, err, nil, nil
		}
		logger.Info("downloading and building DNS block lists")
		var errs []error
		hostnamesLines, ipsLines, errs = dnsConf.BuildBlocked(ctx, client,
			settings.BlockMalicious, settings.BlockAds, settings.BlockSurveillance,
			settings.Unbound.BlockedHostnames, settings.Unbound.BlockedIPs, settings.Unbound.AllowedHostnames,
		)
		for _, err := range errs {
			logger.Warn(err)
		}
		logger.Info("%d hostnames blocked overall", len(hostnamesLines))
		logger.Info("%d IP addresses blocked overall", len(ipsLines))
	}
	logger.Info("generating Unbound configuration")
	if err := dnsConf.MakeUnboundConf(settings.Unbound, hostnamesLines, ipsLines,
		settings.Username, settings.Puid, settings.Pgid); err != nil {
		return oldCtx, oldCancel, err, nil, nil
	}
	newCtx, newCancel = context.WithCancel(ctx)
	oldCancel()
	logger.Info("starting unbound")
	stream, waitFn, err := dnsConf.Start(newCtx, settings.Unbound.VerbosityDetailsLevel)
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
