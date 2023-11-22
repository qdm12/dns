package main

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	_ "github.com/breml/rootcerts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/dns/v2/internal/dns"
	"github.com/qdm12/dns/v2/internal/health"
	"github.com/qdm12/dns/v2/internal/metrics"
	"github.com/qdm12/dns/v2/internal/models"
	"github.com/qdm12/dns/v2/internal/setup"
	"github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/qdm12/goservices"
	"github.com/qdm12/goservices/hooks"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/reader/sources/env"
	"github.com/qdm12/gosettings/reader/sources/flag"
	"github.com/qdm12/gosplash"
	"github.com/qdm12/log"
)

var (
	version string
	created string //nolint:gochecknoglobals
	commit  string //nolint:gochecknoglobals
)

func main() {
	buildInfo := models.BuildInformation{
		Version: version,
		Commit:  commit,
		Created: created,
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	args := os.Args
	logger := log.New()
	settingsSources := []reader.Source{
		flag.New(os.Args),
		env.New(os.Environ()),
	}
	settingsReader := reader.New(reader.Settings{
		Sources: settingsSources,
	})

	errorCh := make(chan error)
	go func() {
		errorCh <- _main(ctx, buildInfo, args, logger, settingsReader)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Caught OS signal, shutting down\n")
		stop()
	case err := <-errorCh:
		if err == nil { // expected exit such as healthcheck
			os.Exit(0)
		}
		logger.Error(err.Error())
		os.Exit(1)
	}

	const shutdownGracePeriod = 5 * time.Second
	timer := time.NewTimer(shutdownGracePeriod)
	select {
	case <-errorCh:
		timer.Stop()
		logger.Info("Shutdown successful")
		os.Exit(0)
	case <-timer.C:
		logger.Warn("Shutdown timed out")
		os.Exit(1)
	}
}

type Logger interface {
	Patch(options ...log.Option)
	New(options ...log.Option) *log.Logger
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(s string)
}

func _main(ctx context.Context, buildInfo models.BuildInformation, //nolint:cyclop
	args []string, logger Logger, settingsReader *reader.Reader) error {
	if health.IsClientMode(args) {
		// Running the program in a separate instance through the Docker
		// built-in healthcheck, in an ephemeral fashion to query the
		// long running instance of the program about its status
		client := health.NewClient()
		return client.Query(ctx)
	}

	initialDisplay(buildInfo)

	var settings config.Settings
	err := settings.Read(settingsReader, logger)
	if err != nil {
		return fmt.Errorf("reading settings: %w", err)
	}
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	logger.Patch(log.SetLevel(*settings.Log.Level))
	logger.Info(settings.String())

	internalDNSSettings := nameserver.SettingsInternalDNS{
		IP: netip.AddrFrom4([4]byte{127, 0, 0, 1}),
	}
	logger.Info("using DNS address " + internalDNSSettings.IP.String() + " internally")
	nameserver.UseDNSInternally(internalDNSSettings) // use the DoT/DoH server

	// Setup health server
	const healthServerAddr = "127.0.0.1:9999"
	healthServerLogger := logger.New(log.SetComponent("health server"))
	healthServer, err := health.NewServer(healthServerAddr, healthServerLogger, health.IsHealthy)
	if err != nil {
		return fmt.Errorf("creating health server: %w", err)
	}

	// Setup DNS loop
	dnsLogger := logger.New(log.SetComponent("DNS server loop"))
	const clientTimeout = 15 * time.Second
	client := &http.Client{Timeout: clientTimeout}
	blockBuilder, err := setup.BuildBlockBuilder(settings.Block, client)
	if err != nil {
		return fmt.Errorf("block builder: %w", err)
	}

	prometheusRegistry := prometheus.NewRegistry()
	cacheMetrics, err := setup.BuildCacheMetrics(settings.Metrics, prometheusRegistry)
	if err != nil {
		return fmt.Errorf("cache metrics: %w", err)
	}
	cache, err := setup.BuildCache(settings.Cache, cacheMetrics) // share the same cache across DNS server restarts
	if err != nil {
		return fmt.Errorf("cache: %w", err)
	}

	dnsLoop, err := dns.New(settings, dnsLogger, blockBuilder, cache, prometheusRegistry)
	if err != nil {
		return fmt.Errorf("creating DNS loop: %w", err)
	}

	// Setup metrics server
	metricsServer, err := metrics.New(settings.Metrics, logger, prometheusRegistry)
	if err != nil {
		return fmt.Errorf("creating metrics server: %w", err)
	}

	hooksLogger := logger.New(log.SetComponent("services"))
	hooks := hooks.NewWithLog(hooksLogger)
	sequenceSettings := goservices.SequenceSettings{
		ServicesStart: []goservices.Service{dnsLoop, metricsServer, healthServer},
		ServicesStop:  []goservices.Service{metricsServer, healthServer, dnsLoop},
		Hooks:         hooks,
	}
	sequence, err := goservices.NewSequence(sequenceSettings)
	if err != nil {
		return fmt.Errorf("creating services sequence: %w", err)
	}

	runError, err := sequence.Start(ctx)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		err = sequence.Stop()
		if err != nil {
			return fmt.Errorf("stopping services: %w", err)
		}
		return nil
	case err = <-runError:
		return err
	}
}

func initialDisplay(buildInfo models.BuildInformation) {
	announcementExp, err := time.Parse(time.RFC3339, "2021-11-20T00:00:00Z")
	if err != nil {
		panic(err)
	}
	splashSettings := gosplash.Settings{
		User:         "qdm12",
		Repository:   "dns",
		Emails:       []string{"quentin.mcgaw@gmail.com"},
		Version:      buildInfo.Version,
		Commit:       buildInfo.Commit,
		BuildDate:    buildInfo.Created,
		Announcement: "Check out qmcgaw/dns:v2.0.0-beta",
		AnnounceExp:  announcementExp,
		// Sponsor information
		PaypalUser:    "qmcgaw",
		GithubSponsor: "qdm12",
	}
	for _, line := range gosplash.MakeLines(splashSettings) {
		fmt.Println(line)
	}
}
