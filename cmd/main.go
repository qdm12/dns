package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/cloudflare-dns-server/internal/dns"
	"github.com/qdm12/cloudflare-dns-server/internal/env"
	"github.com/qdm12/cloudflare-dns-server/internal/healthcheck"
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/cloudflare-dns-server/internal/params"
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
	e := env.New(logger)
	client := network.NewClient(15 * time.Second)
	// Create configurators
	fileManager := files.NewFileManager()
	dnsConf := dns.NewConfigurator(logger, client, fileManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	streamMerger := command.NewStreamMerger(ctx)

	e.PrintVersion(ctx, "Unbound", dnsConf.Version)
	settings := models.Settings{}
	settings.Providers, err = paramsReader.GetProviders()
	e.FatalOnError(err)
	settings.PrivateAddresses = paramsReader.GetPrivateAddresses()
	settings.ListeningPort, err = paramsReader.GetListeningPort()
	e.FatalOnError(err)
	settings.Caching, err = paramsReader.GetCaching()
	e.FatalOnError(err)
	settings.VerbosityLevel, err = paramsReader.GetVerbosity()
	e.FatalOnError(err)
	settings.VerbosityDetailsLevel, err = paramsReader.GetVerbosityDetails()
	e.FatalOnError(err)
	settings.ValidationLogLevel, err = paramsReader.GetValidationLogLevel()
	e.FatalOnError(err)
	settings.BlockMalicious, err = paramsReader.GetMaliciousBlocking()
	e.FatalOnError(err)
	settings.BlockSurveillance, err = paramsReader.GetSurveillanceBlocking()
	e.FatalOnError(err)
	settings.BlockAds, err = paramsReader.GetAdsBlocking()
	e.FatalOnError(err)
	settings.BlockedHostnames, err = paramsReader.GetBlockedHostnames()
	e.FatalOnError(err)
	settings.BlockedIPs, err = paramsReader.GetBlockedIPs()
	e.FatalOnError(err)
	settings.AllowedHostnames, err = paramsReader.GetUnblockedHostnames()
	e.FatalOnError(err)
	settings.CheckUnbound, err = paramsReader.GetCheckUnbound()
	e.FatalOnError(err)
	settings.IPv4, err = paramsReader.GetIPv4()
	e.FatalOnError(err)
	settings.IPv6, err = paramsReader.GetIPv6()
	e.FatalOnError(err)
	logger.Info("Settings summary:\n" + settings.String())

	go func() {
		err = streamMerger.CollectLines(func(line string) { logger.Info(line) })
		e.FatalOnError(err)
	}()

	initialDNSToUse := constants.ProviderMapping()[settings.Providers[0]]
	dnsConf.UseDNSInternally(initialDNSToUse.IPs[0])
	err = dnsConf.DownloadRootHints()
	e.FatalOnError(err)
	err = dnsConf.DownloadRootKey()
	e.FatalOnError(err)
	err = dnsConf.MakeUnboundConf(settings)
	e.FatalOnError(err)
	stream, wait, err := dnsConf.Start(ctx, settings.VerbosityDetailsLevel)
	e.FatalOnError(err)
	go streamMerger.Merge(stream, command.MergeName("unbound"))
	dnsConf.UseDNSInternally(net.IP{127, 0, 0, 1})
	e.FatalOnError(err)
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
