package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	_ "github.com/breml/rootcerts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config/settings"
	"github.com/qdm12/dns/v2/internal/config/sources/env"
	"github.com/qdm12/dns/v2/internal/health"
	"github.com/qdm12/dns/v2/internal/metrics"
	"github.com/qdm12/dns/v2/internal/models"
	"github.com/qdm12/dns/v2/internal/setup"
	"github.com/qdm12/dns/v2/pkg/blockbuilder"
	"github.com/qdm12/dns/v2/pkg/cache"
	"github.com/qdm12/dns/v2/pkg/check"
	"github.com/qdm12/dns/v2/pkg/filter/mapfilter"
	pkglog "github.com/qdm12/dns/v2/pkg/log"
	"github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/qdm12/goshutdown"
	"github.com/qdm12/gosplash"
	"github.com/qdm12/log"
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
	logger := log.New()
	settingsSource := env.New(logger)

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, buildInfo, args, logger, settingsSource)
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

type Logger interface {
	Patch(options ...log.Option)
	New(options ...log.Option) *log.Logger
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

type SettingsSource interface {
	Read() (settings settings.Settings, err error)
}

func _main(ctx context.Context, buildInfo models.BuildInformation,
	args []string, logger Logger, settingsSource SettingsSource) error {
	if health.IsClientMode(args) {
		// Running the program in a separate instance through the Docker
		// built-in healthcheck, in an ephemeral fashion to query the
		// long running instance of the program about its status
		client := health.NewClient()
		return client.Query(ctx)
	}

	announcementExp, err := time.Parse(time.RFC3339, "2021-11-20T00:00:00Z")
	if err != nil {
		return err
	}
	splashSettings := gosplash.Settings{
		User:         "qdm12",
		Repository:   "dns",
		Emails:       []string{"quentin.mcgaw@gmail.com"},
		Version:      buildInfo.Version,
		Commit:       buildInfo.Commit,
		BuildDate:    buildInfo.BuildDate,
		Announcement: "Check out qmcgaw/dns:v2.0.0-beta",
		AnnounceExp:  announcementExp,
		// Sponsor information
		PaypalUser:    "qmcgaw",
		GithubSponsor: "qdm12",
	}
	for _, line := range gosplash.MakeLines(splashSettings) {
		fmt.Println(line)
	}

	const clientTimeout = 15 * time.Second
	client := &http.Client{Timeout: clientTimeout}

	settings, err := settingsSource.Read()
	if err != nil {
		return fmt.Errorf("reading environment variables: %w", err)
	}
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	logger.Patch(log.SetLevel(*settings.Log.Level))
	logger.Info(settings.String())

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.New(log.SetComponent("healthcheck server")),
		health.IsHealthy)
	healthServerHandler, healthServerCtx, healthServerDone := goshutdown.NewGoRoutineHandler(
		"health server", goshutdown.GoRoutineSettings{})
	go healthServer.Run(healthServerCtx, healthServerDone)

	internalDNSSettings := nameserver.SettingsInternalDNS{
		IP: net.IP{127, 0, 0, 1},
	}
	logger.Info("using DNS address " + internalDNSSettings.IP.String() + " internally")
	nameserver.UseDNSInternally(internalDNSSettings) // use the DoT/DoH server

	blockBuilder := setup.BuildBlockBuilder(settings.Block, client)

	prometheusRegistry := prometheus.NewRegistry()

	cacheMetrics, err := setup.CacheMetrics(settings.Metrics, prometheusRegistry)
	if err != nil {
		return fmt.Errorf("cache metrics: %w", err)
	}

	// Use the same cache across DNS server restarts
	cache := setup.BuildCache(settings.Cache, cacheMetrics)

	dnsServerHandler, dnsServerCtx, dnsServerDone := goshutdown.NewGoRoutineHandler(
		"dns server", goshutdown.GoRoutineSettings{})
	crashed := make(chan error)
	go runLoop(dnsServerCtx, dnsServerDone, crashed, settings,
		logger, blockBuilder, cache, prometheusRegistry)

	metricsServer := metrics.Setup(settings.Metrics, logger, prometheusRegistry)
	metricsServerHandler, metricsServerCtx, metricsServerDone := goshutdown.NewGoRoutineHandler(
		"metrics server", goshutdown.GoRoutineSettings{})
	go metricsServer.Run(metricsServerCtx, metricsServerDone)

	group := goshutdown.NewGroupHandler("", goshutdown.GroupSettings{})
	group.Add(healthServerHandler, metricsServerHandler, dnsServerHandler)

	select {
	case <-ctx.Done():
	case err := <-crashed:
		logger.Error(err.Error())
	}

	return group.Shutdown(context.Background()) //nolint:contextcheck
}

//nolint:cyclop,gocognit
func runLoop(ctx context.Context, dnsServerDone chan<- struct{},
	crashed chan<- error, settings settings.Settings,
	logger pkglog.Logger, blockBuilder blockbuilder.Interface,
	cache cache.Interface, prometheusRegistry prometheus.Registerer) {
	defer close(dnsServerDone)

	timer := time.NewTimer(time.Hour)

	firstRun := true

	var (
		serverCtx    context.Context
		serverCancel context.CancelFunc
		waitError    chan error
	)

	for {
		timer.Stop()
		if *settings.UpdatePeriod > 0 {
			timer.Reset(*settings.UpdatePeriod)
		}

		filterMetrics, err := setup.FilterMetrics(settings.Metrics, prometheusRegistry)
		if err != nil {
			serverCancel()
			crashed <- err
			return
		}
		filterSettings := mapfilter.Settings{
			Metrics: filterMetrics,
		}
		if !firstRun {
			logger.Info("downloading and building DNS block lists")
			result := blockBuilder.BuildAll(ctx)
			for _, err := range result.Errors {
				logger.Warn(err.Error())
			}
			logger.Info(fmt.Sprint(len(result.BlockedHostnames)) + " hostnames blocked overall")
			logger.Info(fmt.Sprint(len(result.BlockedIPs)) + " IP addresses blocked overall")
			logger.Info(fmt.Sprint(len(result.BlockedIPPrefixes)) + " IP networks blocked overall")
			filterSettings.Update.IPs = result.BlockedIPs
			filterSettings.Update.IPPrefixes = result.BlockedIPPrefixes
			filterSettings.Update.BlockHostnames(result.BlockedHostnames)

			serverCancel()
			<-waitError
			close(waitError)
		}

		filter := mapfilter.New(filterSettings)

		serverCtx, serverCancel = context.WithCancel(ctx)

		logMiddlewareSettings, err := setup.MiddlewareLogger(settings.MiddlewareLog)
		if err != nil {
			crashed <- err
			serverCancel()
			return
		}

		server, err := setup.DNS(serverCtx, settings, cache,
			filter, logger, prometheusRegistry, logMiddlewareSettings)
		if err != nil {
			crashed <- err
			serverCancel()
			return
		}

		logger.Info("starting DNS server")
		waitError = make(chan error)
		go server.Run(serverCtx, waitError)

		if *settings.CheckDNS {
			if err := check.WaitForDNS(ctx, check.Settings{}); err != nil {
				crashed <- err
				serverCancel()
				return
			}
		}

		if firstRun {
			logger.Info("restarting DNS server the first time to get updated files")
			firstRun = false
			continue
		}

		select {
		case <-timer.C:
			logger.Info("planned periodic restart of DNS server")
		case <-ctx.Done():
			logger.Warn("exiting DNS server run loop (" + ctx.Err().Error() + ")")
			if !timer.Stop() {
				<-timer.C
			}
			if err := <-waitError; err != nil {
				logger.Error(err.Error())
			}
			close(waitError)
			serverCancel()
			return

		case waitErr := <-waitError:
			close(waitError)
			if !timer.Stop() {
				<-timer.C
			}
			serverCancel()
			crashed <- waitErr
			return
		}
	}
}
