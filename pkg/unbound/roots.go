package unbound

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func (c *configurator) SetupFiles(ctx context.Context) error {
	if err := c.downloadRootHints(ctx); err != nil {
		return err
	}

	if err := c.downloadRootKeys(ctx); err != nil {
		return err
	}

	return c.createEmptyIncludeConf()
}

func (c *configurator) downloadRootHints(ctx context.Context) error {
	namedRoot, err := c.dnscrypto.DownloadNamedRoot(ctx)
	if err != nil {
		return err
	}

	filepath := filepath.Join(c.unboundEtcDir, rootHints)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := file.Write(namedRoot); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

func (c *configurator) downloadRootKeys(ctx context.Context) error {
	rootAnchorsXML, err := c.dnscrypto.DownloadRootAnchorsXML(ctx)
	if err != nil {
		return err
	}
	rootKeys, err := c.dnscrypto.ConvertRootAnchorsToRootKeys(rootAnchorsXML)
	if err != nil {
		return err
	}

	filepath := filepath.Join(c.unboundEtcDir, rootKey)
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := file.WriteString(strings.Join(rootKeys, "\n")); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}
