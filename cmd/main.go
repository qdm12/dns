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

const (
	uid, gid = 1000, 1000
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

	e.PrintVersion("Unbound", dnsConf.Version)
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
	logger.Info("Settings summary:\n" + settings.String())

	go func() {
		err = streamMerger.CollectLines(func(line string) { logger.Info(line) })
		e.FatalOnError(err)
	}()

	initialDNSToUse := constants.ProviderMapping()[settings.Providers[0]]
	dnsConf.UseDNSInternally(initialDNSToUse.IPs[0])
	err = dnsConf.DownloadRootHints(uid, gid)
	e.FatalOnError(err)
	err = dnsConf.DownloadRootKey(uid, gid)
	e.FatalOnError(err)
	err = dnsConf.MakeUnboundConf(settings, uid, gid)
	e.FatalOnError(err)
	stream, err := dnsConf.Start(settings.VerbosityDetailsLevel)
	e.FatalOnError(err)
	go streamMerger.Merge("unbound", stream)
	dnsConf.UseDNSInternally(net.IP{127, 0, 0, 1})
	e.FatalOnError(err)
	err = dnsConf.WaitForUnbound()
	e.FatalOnError(err)

	signals.WaitForExit(func(signal string) int {
		logger.Warn("Caught OS signal %s, shutting down", signal)
		time.Sleep(100 * time.Millisecond) // wait for other processes to exit
		return 0
	})
}
