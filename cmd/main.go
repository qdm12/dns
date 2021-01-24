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
	go unboundRunLoop(ctx, wg, settings, logger, dnsConf, client, cancel)

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

func unboundRunLoop(ctx context.Context, wg *sync.WaitGroup, settings models.Settings, //nolint:gocognit
	logger logging.Logger, dnsConf unbound.Configurator, client *http.Client, fatal func(),
) {
	defer wg.Done()
	defer logger.Info("unbound loop exited")
	timer := time.NewTimer(time.Hour)

	unboundCtx, unboundCancel := context.WithCancel(ctx) //nolint:ineffassign,staticcheck

	firstRun := true
	restart := false

	var waitError chan error
	var stdoutLines, stderrLines chan string

	for ctx.Err() == nil {
		timer.Stop()
		if settings.UpdatePeriod > 0 {
			timer.Reset(settings.UpdatePeriod)
		}

		var hostnamesLines, ipsLines []string
		if !firstRun {
			logger.Info("downloading DNSSEC root hints and named root")
			if err := dnsConf.SetupFiles(ctx); err != nil {
				logAndWait(ctx, logger, err)
				continue
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
			logAndWait(ctx, logger, err)
			continue
		}

		if restart {
			unboundCancel()
			<-waitError
			close(waitError)
			close(stdoutLines)
			close(stderrLines)
		}
		unboundCtx, unboundCancel = context.WithCancel(ctx)

		logger.Info("starting unbound")
		stdoutLines, stderrLines, waitError, err := dnsConf.Start(unboundCtx, settings.Unbound.VerbosityDetailsLevel)
		if err != nil {
			logger.Error(err)
			unboundCancel()
			fatal()
		}

		go logUnboundStreams(logger, stdoutLines, stderrLines)

		if settings.CheckUnbound {
			if err := dnsConf.WaitForUnbound(ctx); err != nil {
				logger.Error(err)
				unboundCancel()
				fatal()
			}
		}

		if firstRun {
			logger.Info("restarting Unbound the first time to get updated files")
			firstRun = false
			continue
		}

		select {
		case <-timer.C:
			logger.Info("planned restart of unbound")
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			logger.Warn("context canceled: exiting unbound run loop")
		case waitErr := <-waitError:
			close(waitError)
			close(stdoutLines)
			close(stderrLines)
			unboundCancel()
			if !timer.Stop() {
				<-timer.C
			}
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

func logUnboundStreams(logger logging.Logger, stdout, stderr <-chan string) {
	var line string
	var ok bool
	for {
		select {
		case line, ok = <-stdout:
		case line, ok = <-stderr:
		}
		if !ok {
			fmt.Println("EXITIIIITTT")
			return
		}
		logger.Info(line)
	}
}
