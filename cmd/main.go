package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
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
	customOS "github.com/qdm12/golibs/os"
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
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	args := os.Args
	logger := logging.New(logging.StdLog)
	paramsReader := params.NewParamsReader(logger)
	osIntf := customOS.New()

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, buildInfo, args, logger, paramsReader, osIntf)
	}()

	signalsCh := make(chan os.Signal, 1)
	signal.Notify(signalsCh,
		syscall.SIGINT,
		syscall.SIGTERM,
		os.Interrupt,
	)

	select {
	case <-ctx.Done():
		logger.Warn("Caught OS signal, shutting down\n")
		stop()
	case err := <-errorCh:
		close(errorCh)
		if err == nil { // expected exit such as healthcheck
			os.Exit(0)
		}
		logger.Error(err)
	}

	const shutdownGracePeriod = 5 * time.Second
	timer := time.NewTimer(shutdownGracePeriod)
	select {
	case <-errorCh:
		if !timer.Stop() {
			<-timer.C
		}
		logger.Info("Shutdown successful")
	case <-timer.C:
		logger.Warn("Shutdown timed out")
	}

	os.Exit(1)
}

func _main(ctx context.Context, buildInfo models.BuildInformation,
	args []string, logger logging.Logger, paramsReader params.Reader,
	os customOS.OS) error {
	if health.IsClientMode(args) {
		// Running the program in a separate instance through the Docker
		// built-in healthcheck, in an ephemeral fashion to query the
		// long running instance of the program about its status
		client := health.NewClient()
		if err := client.Query(ctx); err != nil {
			return err
		}
		return nil
	}
	fmt.Println(splash.Splash(buildInfo))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
			return err
		}
		return nil
	}

	version, err := dnsConf.Version(ctx)
	if err != nil {
		return err
	}
	logger.Info("Unbound version: %s", version)

	settings, err := settings.GetSettings(paramsReader)
	if err != nil {
		return err
	}
	logger.Info("Settings summary:\n" + settings.String())

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	crashed := make(chan error)

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.NewChild(logging.SetPrefix("healthcheck server: ")),
		health.IsHealthy)
	wg.Add(1)
	go healthServer.Run(ctx, wg)

	localIP := net.IP{127, 0, 0, 1}
	logger.Info("using DNS address %s internally", localIP.String())
	dnsConf.UseDNSInternally(localIP) // use Unbound
	wg.Add(1)
	go unboundRunLoop(ctx, wg, settings, logger, dnsConf, client, crashed)

	select {
	case <-ctx.Done():
	case err = <-crashed:
		cancel()
	}
	wg.Wait()
	return err
}

func unboundRunLoop(ctx context.Context, wg *sync.WaitGroup, settings models.Settings, //nolint:gocognit
	logger logging.Logger, dnsConf unbound.Configurator, client *http.Client, crashed chan<- error,
) {
	defer wg.Done()
	defer logger.Info("unbound loop exited")
	timer := time.NewTimer(time.Hour)

	firstRun := true
	restart := false

	var (
		unboundCtx               context.Context
		unboundCancel            context.CancelFunc
		waitError                chan error
		stdoutLines, stderrLines chan string
	)

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
			crashed <- err
			break
		}

		go logUnboundStreams(logger, stdoutLines, stderrLines)

		if settings.CheckUnbound {
			if err := dnsConf.WaitForUnbound(ctx); err != nil {
				crashed <- err
				break
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
			if !timer.Stop() {
				<-timer.C
			}
			crashed <- waitErr
			break
		}
	}
	unboundCancel()
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
