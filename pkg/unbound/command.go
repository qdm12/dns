package unbound

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func (c *configurator) Start(ctx context.Context, verbosityDetailsLevel uint8) (
	stdoutLines, stderrLines chan string, waitError chan error, err error) {
	configFilepath := filepath.Join(c.unboundEtcDir, unboundConfigFilename)
	args := []string{"-d", "-c", configFilepath}
	if verbosityDetailsLevel > 0 {
		args = append(args, "-"+strings.Repeat("v", int(verbosityDetailsLevel)))
	}

	cmd := exec.CommandContext(ctx, c.unboundPath, args...) //nolint:gosec
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return c.cmder.Start(cmd)
}

var ErrVersionNotFound = errors.New("unbound version not found")

func (c *configurator) Version(ctx context.Context) (version string, err error) {
	cmd := exec.CommandContext(ctx, c.unboundPath, "-V") //nolint:gosec

	output, err := c.cmder.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("unbound version: %w", err)
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Version ") {
			words := strings.Fields(line)
			const minWords = 2
			if len(words) < minWords {
				continue
			}
			version = words[1]
		}
	}
	if version == "" {
		return "", fmt.Errorf("%w: %s", ErrVersionNotFound, output)
	}
	return version, nil
}
