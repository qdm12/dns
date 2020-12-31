package dns

import (
	"context"
	"io"
	"net"

	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/files"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/network"
)

type Configurator interface {
	DownloadRootHints(ctx context.Context, client network.Client) error
	DownloadRootKey(ctx context.Context, client network.Client) error
	MakeUnboundConf(settings models.Settings, hostnamesLines, ipsLines []string) (err error)
	UseDNSInternally(IP net.IP)
	Start(ctx context.Context, logLevel uint8) (stdout io.ReadCloser, wait func() error, err error)
	WaitForUnbound(ctx context.Context) (err error)
	Version(ctx context.Context) (version string, err error)
	BuildBlocked(ctx context.Context, client network.Client,
		blockMalicious, blockAds, blockSurveillance bool,
		blockedHostnames, blockedIPs, allowedHostnames []string) (
		hostnamesLines, ipsLines []string)
}

type configurator struct {
	logger      logging.Logger
	fileManager files.FileManager
	commander   command.Commander
	resolver    *net.Resolver
}

func NewConfigurator(logger logging.Logger, fileManager files.FileManager) Configurator {
	return &configurator{
		logger:      logger,
		fileManager: fileManager,
		commander:   command.NewCommander(),
		resolver:    net.DefaultResolver,
	}
}
