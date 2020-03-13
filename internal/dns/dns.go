package dns

import (
	"io"
	"net"

	"github.com/qdm12/cloudflare-dns-server/internal/models"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/files"
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/network"
)

type Configurator interface {
	DownloadRootHints() error
	DownloadRootKey() error
	MakeUnboundConf(settings models.Settings) (err error)
	UseDNSInternally(IP net.IP)
	Start(logLevel uint8) (stdout io.ReadCloser, err error)
	WaitForUnbound() (err error)
	Version() (version string, err error)
}

type configurator struct {
	logger      logging.Logger
	client      network.Client
	fileManager files.FileManager
	commander   command.Commander
	lookupIP    func(host string) ([]net.IP, error)
}

func NewConfigurator(logger logging.Logger, client network.Client, fileManager files.FileManager) Configurator {
	return &configurator{
		logger:      logger,
		client:      client,
		fileManager: fileManager,
		commander:   command.NewCommander(),
		lookupIP:    net.LookupIP,
	}
}
