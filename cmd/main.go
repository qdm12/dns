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
	"github.com/qdm12/dns/pkg/doh"
	"github.com/qdm12/dns/pkg/dot"
	"github.com/qdm12/dns/pkg/nameserver"
	"github.com/qdm12/golibs/logging"
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

	settings, err := configReader.ReadSettings()
	if err != nil {
		return err
	}
	logger = logger.NewChild(logging.Settings{
		Level: settings.LogLevel,
	})
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
	nameserver.UseDNSInternally(localIP) // use the DoT/DoH server
	wg.Add(1)
	go runLoop(ctx, wg, settings, logger, client, crashed)

	select {
	case <-ctx.Done():
	case err = <-crashed:
		cancel()
	}
	wg.Wait()
	return err
}

func runLoop(ctx context.Context, wg *sync.WaitGroup, settings config.Settings,
	logger logging.Logger, client *http.Client, crashed chan<- error,
) {
	defer wg.Done()
	defer logger.Info("unbound loop exited")
	timer := time.NewTimer(time.Hour)

	firstRun := true

	var (
		serverCtx    context.Context
		serverCancel context.CancelFunc
		waitError    chan error
	)

	for ctx.Err() == nil {
		timer.Stop()
		if settings.UpdatePeriod > 0 {
			timer.Reset(settings.UpdatePeriod)
		}

		serverSettings := dot.ServerSettings{
			Resolver: settings.DoT.Resolver,
			Port:     settings.DoT.Port,
			Cache:    settings.DoT.Cache,
		}

		if !firstRun {
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
			serverSettings.Blacklist.IPs = blockedIPs
			serverSettings.Blacklist.IPPrefixes = blockedIPPrefixes
			serverSettings.Blacklist.BlockHostnames(blockedHostnames)
		}

		if !firstRun {
			serverCancel()
			<-waitError
			close(waitError)
		}
		serverCtx, serverCancel = context.WithCancel(ctx)

		var server models.Server
		switch settings.UpstreamType {
		case config.DoT:
			server = dot.NewServer(serverCtx, logger, serverSettings)
		case config.DoH:
			server = doh.NewServer(serverCtx, logger, settings.DoH)
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
			if !timer.Stop() {
				<-timer.C
			}
			logger.Warn("context canceled: exiting DNS server run loop")
		case waitErr := <-waitError:
			close(waitError)
			if !timer.Stop() {
				<-timer.C
			}
			crashed <- waitErr
			serverCancel()
			return
		}
	}
	serverCancel() // for the linter
}
