package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/qdm12/dns/internal/config"
	"github.com/qdm12/dns/internal/health"
	"github.com/qdm12/dns/internal/models"
	"github.com/qdm12/dns/internal/splash"
	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/dns/pkg/check"
	"github.com/qdm12/dns/pkg/nameserver"
	"github.com/qdm12/dns/pkg/unbound"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/updated/pkg/dnscrypto"
)

var (
	version   string
	buildDate string //nolint:gochecknoglobals
	commit    string //nolint:gochecknoglobals
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
	logger := logging.NewParent(logging.Settings{})
	configReader := config.NewReader(logger)

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, buildInfo, args, logger, configReader)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Caught OS signal, shutting down\n")
		stop()
	case err := <-errorCh:
		close(errorCh)
		if err == nil { // expected exit such as healthcheck
			os.Exit(0)
		}
		logger.Error(err.Error())
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
	args []string, logger logging.ParentLogger, configReader config.Reader) error {
	if health.IsClientMode(args) {
		// Running the program in a separate instance through the Docker
		// built-in healthcheck, in an ephemeral fashion to query the
		// long running instance of the program about its status
		client := health.NewClient()
		return client.Query(ctx)
	}
	fmt.Println(splash.Splash(buildInfo))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	const clientTimeout = 15 * time.Second
	client := &http.Client{Timeout: clientTimeout}
	// Create configurators
	dnsCrypto := dnscrypto.New(client, "", "") // TODO checksums for build
	cmder := command.NewCmder()
	const unboundEtcDir = "/unbound"
	const unboundPath = "/unbound/unbound"
	const cacertsPath = "/unbound/ca-certificates.crt"
	dnsConf := unbound.NewConfigurator(logger, cmder, dnsCrypto,
		unboundEtcDir, unboundPath, cacertsPath)

	if len(args) > 1 && args[1] == "build" {
		return dnsConf.SetupFiles(ctx)
	}

	version, err := dnsConf.Version(ctx)
	if err != nil {
		return err
	}
	logger.Info("Unbound version: " + version)

	settings, err := configReader.ReadSettings()
	if err != nil {
		return err
	}
	logger.Info("Settings summary:\n" + settings.String())

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	crashed := make(chan error)

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.NewChild(logging.Settings{Prefix: "healthcheck server: "}),
		health.IsHealthy)
	wg.Add(1)
	go healthServer.Run(ctx, wg)

	localIP := net.IP{127, 0, 0, 1}
	logger.Info("using DNS address " + localIP.String() + " internally")
	nameserver.UseDNSInternally(localIP) // use Unbound
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

func unboundRunLoop(ctx context.Context, wg *sync.WaitGroup, settings config.Settings, //nolint:gocognit
	logger logging.Logger, dnsConf unbound.Configurator, client *http.Client, crashed chan<- error,
) {
	defer wg.Done()
	defer logger.Info("unbound loop exited")
	timer := time.NewTimer(time.Hour)

	firstRun := true

	var (
		unboundCtx               context.Context
		unboundCancel            context.CancelFunc
		waitError                chan error
		stdoutLines, stderrLines chan string
		err                      error
	)

	for ctx.Err() == nil {
		timer.Stop()
		if settings.UpdatePeriod > 0 {
			timer.Reset(settings.UpdatePeriod)
		}

		if !firstRun {
			logger.Info("downloading DNSSEC root hints and named root")
			if err := dnsConf.SetupFiles(ctx); err != nil {
				logAndWait(ctx, logger, err)
				continue
			}
			logger.Info("downloading and building DNS block lists")
			blacklistBuilder := blacklist.NewBuilder(client)
			blockedHostnames, blockedIPs, blockedIPPrefixes, errs :=
				blacklistBuilder.All(ctx, settings.Blacklist)
			for _, err := range errs {
				logger.Warn(err.Error())
			}
			logger.Info(strconv.Itoa(len(blockedHostnames)) + " hostnames blocked overall")
			logger.Info(strconv.Itoa(len(blockedIPs)) + " IP addresses blocked overall")
			logger.Info(strconv.Itoa(len(blockedIPPrefixes)) + " IP networks blocked overall")
			settings.Unbound.Blacklist = blacklist.Settings{
				FqdnHostnames: blockedHostnames,
				IPs:           blockedIPs,
				IPPrefixes:    blockedIPPrefixes,
			}
		}

		logger.Info("generating Unbound configuration")
		if err := dnsConf.MakeUnboundConf(settings.Unbound); err != nil {
			logAndWait(ctx, logger, err)
			continue
		}

		if !firstRun {
			unboundCancel()
			<-waitError
			close(waitError)
			close(stdoutLines)
			close(stderrLines)
		}
		unboundCtx, unboundCancel = context.WithCancel(ctx)

		logger.Info("starting unbound")
		stdoutLines, stderrLines, waitError, err = dnsConf.Start(unboundCtx, settings.Unbound.VerbosityDetailsLevel)
		if err != nil {
			crashed <- err
			break
		}

		go logUnboundStreams(logger, stdoutLines, stderrLines)

		if settings.CheckDNS {
			if err := check.WaitForDNS(ctx, net.DefaultResolver); err != nil {
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
			unboundCancel()
			return
		}
	}
	unboundCancel()
}

func logAndWait(ctx context.Context, logger logging.Logger, err error) {
	const wait = 10 * time.Second
	logger.Error(err.Error() + ", retrying in " + wait.String())
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
			return
		}
		logger.Info(line)
	}
}
