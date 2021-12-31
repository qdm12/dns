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
	"github.com/qdm12/dns/internal/cache"
	"github.com/qdm12/dns/internal/config"
	"github.com/qdm12/dns/internal/health"
	"github.com/qdm12/dns/internal/metrics"
	"github.com/qdm12/dns/internal/models"
	"github.com/qdm12/dns/pkg/blockbuilder"
	"github.com/qdm12/dns/pkg/check"
	"github.com/qdm12/dns/pkg/doh"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/dns/pkg/filter/mapfilter"
	"github.com/qdm12/dns/pkg/log"
	"github.com/qdm12/dns/pkg/nameserver"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/goshutdown"
	"github.com/qdm12/gosplash"
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
	logger := logging.New(logging.Settings{})
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
	args []string, logger logging.ParentLogger, configReader config.SettingsReader) error {
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

	settings, err := configReader.ReadSettings()
	if err != nil {
		return err
	}
	logger.PatchLevel(settings.Log.Level)
	logger.Info(settings.String())

	const healthServerAddr = "127.0.0.1:9999"
	healthServer := health.NewServer(healthServerAddr,
		logger.NewChild(logging.Settings{Prefix: "healthcheck server: "}),
		health.IsHealthy)
	healthServerHandler, healthServerCtx, healthServerDone := goshutdown.NewGoRoutineHandler(
		"health server", goshutdown.GoRoutineSettings{})
	go healthServer.Run(healthServerCtx, healthServerDone)

	internalDNSSettings := nameserver.SettingsInternalDNS{
		IP: net.IP{127, 0, 0, 1},
	}
	logger.Info("using DNS address " + internalDNSSettings.IP.String() + " internally")
	nameserver.UseDNSInternally(internalDNSSettings) // use the DoT/DoH server

	settings.PatchLogger(logger)

	metricsServer, err := metrics.Setup(&settings, logger)
	if err != nil {
		return err
	}

	// Use the same cache across DNS server restarts
	cache.Setup(&settings)

	blockBuilderSettings := blockbuilder.Settings{Client: client}
	blockBuilder := blockbuilder.New(blockBuilderSettings)

	dnsServerHandler, dnsServerCtx, dnsServerDone := goshutdown.NewGoRoutineHandler(
		"dns server", goshutdown.GoRoutineSettings{})
	crashed := make(chan error)
	go runLoop(dnsServerCtx, dnsServerDone, crashed, settings, logger, blockBuilder)

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

	return group.Shutdown(context.Background())
}

func runLoop(ctx context.Context, dnsServerDone chan<- struct{},
	crashed chan<- error, settings config.Settings,
	logger log.Logger, blockBuilder blockbuilder.Interface) {
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
		if settings.UpdatePeriod > 0 {
			timer.Reset(settings.UpdatePeriod)
		}

		if !firstRun {
			logger.Info("downloading and building DNS block lists")
			result := blockBuilder.BuildAll(ctx, settings.BlockBuilder)
			for _, err := range result.Errors {
				logger.Warn(err.Error())
			}
			logger.Info(fmt.Sprint(len(result.BlockedHostnames)) + " hostnames blocked overall")
			logger.Info(fmt.Sprint(len(result.BlockedIPs)) + " IP addresses blocked overall")
			logger.Info(fmt.Sprint(len(result.BlockedIPPrefixes)) + " IP networks blocked overall")
			settings.Filter.Update.IPs = result.BlockedIPs
			settings.Filter.Update.IPPrefixes = result.BlockedIPPrefixes
			settings.Filter.Update.BlockHostnames(result.BlockedHostnames)

			serverCancel()
			<-waitError
			close(waitError)
		}

		filter := mapfilter.New(settings.Filter)
		settings.PatchFilter(filter)

		serverCtx, serverCancel = context.WithCancel(ctx)

		var server models.Server
		switch settings.UpstreamType {
		case config.DoT:
			server = dot.NewServer(serverCtx, settings.DoT)
		case config.DoH:
			server = doh.NewServer(serverCtx, settings.DoH)
		}

		logger.Info("starting DNS server")
		waitError = make(chan error)
		go server.Run(serverCtx, waitError)

		if settings.CheckDNS {
			if err := check.WaitForDNS(ctx, net.DefaultResolver); err != nil {
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
