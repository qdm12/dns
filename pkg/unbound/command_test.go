package unbound

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qdm12/golibs/command"
	"github.com/qdm12/golibs/command/mock_command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Start(t *testing.T) {
	t.Parallel()
	mockCtrl := gomock.NewController(t)
	const unboundEtcDir = "/unbound"
	const unboundPath = "/usr/sbin/unbound"

	ctx := context.Background()
	commander := mock_command.NewMockCommander(mockCtrl)
	cmd := exec.CommandContext(ctx, unboundPath, "-d", "-c", "/unbound/unbound.conf", "-vv")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	commander.EXPECT().Start(cmd).
		DoAndReturn(func(cmd command.Cmd) (
			stdoutLines, stderrLines chan string, waitError chan error, err error) {
			stdoutLines = make(chan string)
			stderrLines = make(chan string)
			waitError = make(chan error, 1) // buffered so it runs in the same goroutine
			waitError <- nil
			return stdoutLines, stderrLines, waitError, nil
		})

	c := &configurator{
		commander:     commander,
		unboundEtcDir: unboundEtcDir,
		unboundPath:   unboundPath,
	}

	stdoutLines, stderrLines, waitError, err := c.Start(ctx, 2)

	assert.NoError(t, err)
	<-waitError
	close(stdoutLines)
	close(stderrLines)
	close(waitError)
}

func Test_Version(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		runOutput string
		runErr    error
		version   string
		err       error
	}{
		"no data": {
			err: fmt.Errorf(`unbound version was not found in ""`),
		},
		"2 lines with version": {
			runOutput: "Version  \nVersion 1.0-a hello\n",
			version:   "1.0-a",
		},
		"run error": {
			runErr: fmt.Errorf("error"),
			err:    fmt.Errorf("unbound version: error"),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			commander := mock_command.NewMockCommander(mockCtrl)
			ctx := context.Background()

			const unboundEtcDir = "/unbound"
			const unboundPath = "/usr/sbin/unbound"

			cmd := exec.CommandContext(ctx, unboundPath, "-V")

			commander.EXPECT().Run(cmd).Return(tc.runOutput, tc.runErr)
			c := &configurator{
				commander:     commander,
				unboundEtcDir: unboundEtcDir,
				unboundPath:   unboundPath,
			}
			version, err := c.Version(ctx)
			if tc.err != nil {
				require.Error(t, err)
				assert.Equal(t, tc.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.version, version)
		})
	}
}
