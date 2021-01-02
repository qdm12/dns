package unbound

import (
	"os"
	"path/filepath"
)

const includeConfFilename = "include.conf"

func (c *configurator) createEmptyIncludeConf() error {
	filepath := filepath.Join(c.unboundEtcDir, includeConfFilename)
	file, err := c.openFile(filepath, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}
