package dns

import (
	"context"
	"os"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/golibs/files"
)

func (c *configurator) SetupFiles(ctx context.Context) error {
	if err := c.downloadRootHints(ctx); err != nil {
		return err
	}

	if err := c.downloadRootKeys(ctx); err != nil {
		return err
	}

	return nil
}

func (c *configurator) downloadRootHints(ctx context.Context) error {
	c.logger.Info("downloading root hints")
	namedRoot, err := c.dnscrypto.GetNamedRoot(ctx)
	if err != nil {
		return err
	}
	const userWritePerm os.FileMode = 0600
	return c.fileManager.WriteToFile(
		string(constants.RootHints),
		namedRoot,
		files.Permissions(userWritePerm))
}

func (c *configurator) downloadRootKeys(ctx context.Context) error {
	c.logger.Info("downloading root keys")
	rootAnchorsXML, err := c.dnscrypto.GetRootAnchorsXML(ctx)
	if err != nil {
		return err
	}
	rootKeys, err := c.dnscrypto.ConvertRootAnchorsToRootKeys(rootAnchorsXML)
	if err != nil {
		return err
	}
	const userWritePerm os.FileMode = 0600
	return c.fileManager.WriteLinesToFile(
		string(constants.RootKey),
		rootKeys,
		files.Permissions(userWritePerm))
}
