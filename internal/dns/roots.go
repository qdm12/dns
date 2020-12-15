package dns

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/golibs/files"
)

func (c *configurator) DownloadRootHints(ctx context.Context) error {
	c.logger.Info("downloading root hints from %s", constants.NamedRootURL)
	content, status, err := c.client.Get(ctx, string(constants.NamedRootURL))
	if err != nil {
		return err
	} else if status != http.StatusOK {
		return fmt.Errorf("HTTP status code is %d for %s", status, constants.NamedRootURL)
	}
	const userWritePerm os.FileMode = 0600
	return c.fileManager.WriteToFile(
		string(constants.RootHints),
		content,
		files.Permissions(userWritePerm))
}

func (c *configurator) DownloadRootKey(ctx context.Context) error {
	c.logger.Info("downloading root key from %s", constants.RootKeyURL)
	content, status, err := c.client.Get(ctx, string(constants.RootKeyURL))
	if err != nil {
		return err
	} else if status != http.StatusOK {
		return fmt.Errorf("HTTP status code is %d for %s", status, constants.RootKeyURL)
	}
	const userWritePerm os.FileMode = 0600
	return c.fileManager.WriteToFile(
		string(constants.RootKey),
		content,
		files.Permissions(userWritePerm))
}
