package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/cloudflare-dns-server/internal/dns"
	"github.com/qdm12/cloudflare-dns-server/internal/healthcheck"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
	"github.com/qdm12/cloudflare-dns-server/internal/settings"
	"github.com/qdm12/cloudflare-dns-server/internal/splash"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/files"
	libhealthcheck "github.com/qdm12/golibs/healthcheck"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/network"
	"github.com/qdm12/golibs/signals"
)

func main() {
	logger, err := logging.NewLogger(logging.ConsoleEncoding, logging.InfoLevel, -1)
	if err != nil {
		panic(err)
	}
	if libhealthcheck.Mode(os.Args) {
		if err := healthcheck.Healthcheck(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	paramsReader := params.NewParamsReader(logger)
	fmt.Println(splash.Splash(paramsReader))
	client := network.NewClient(15 * time.Second)
	// Create configurators
	fileManager := files.NewFileManager()
	dnsConf := dns.NewConfigurator(logger, client, fileManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	streamMerger := command.NewStreamMerger()

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

	go streamMerger.CollectLines(ctx,
		func(line string) { logger.Info(line) },
		func(err error) { logger.Error(err) })

	initialDNSToUse := constants.ProviderMapping()[settings.Providers[0]]
	dnsConf.UseDNSInternally(initialDNSToUse.IPs[0])
	err = dnsConf.DownloadRootHints()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = dnsConf.DownloadRootKey()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	err = dnsConf.MakeUnboundConf(settings)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	stream, wait, err := dnsConf.Start(ctx, settings.VerbosityDetailsLevel)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	go streamMerger.Merge(ctx, stream, command.MergeName("unbound"))
	dnsConf.UseDNSInternally(net.IP{127, 0, 0, 1})
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	if settings.CheckUnbound {
		if err := dnsConf.WaitForUnbound(); err != nil {
			logger.Warn(err)
		}
	}
	signals.WaitForExit(func(signal string) int {
		logger.Warn("Caught OS signal %s, shutting down", signal)
		cancel()
		if err := wait(); err != nil {
			logger.Warn(err)
		}
		return 0
	})
}
