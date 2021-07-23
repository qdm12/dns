package nameserver

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UseDNSSystemWide(t *testing.T) {
	t.Parallel()

	t.Run("file does not exist", func(t *testing.T) {
		t.Parallel()

		dirPath, err := os.MkdirTemp("", "")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dirPath)
			require.NoError(t, err)
		}()

		resolvConfPath := filepath.Join(dirPath, "resolv.conf")
		ip := net.IP{1, 1, 1, 1}
		const keepNameserver = false

		err = UseDNSSystemWide(resolvConfPath, ip, keepNameserver)

		require.Error(t, err)
		assert.Equal(t, "open "+resolvConfPath+": no such file or directory", err.Error())
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		file, err := os.CreateTemp("", "")
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		resolvConfPath := file.Name()

		defer func() {
			err := os.Remove(resolvConfPath)
			require.NoError(t, err)
		}()

		ip := net.IP{1, 1, 1, 1}
		const keepNameserver = false

		err = UseDNSSystemWide(resolvConfPath, ip, keepNameserver)

		require.NoError(t, err)

		file, err = os.Open(resolvConfPath)
		require.NoError(t, err)
		b, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\n", string(b))
	})

	t.Run("preserve nameserver", func(t *testing.T) {
		t.Parallel()

		file, err := os.CreateTemp("", "")
		require.NoError(t, err)
		_, err = io.WriteString(file, "nameserver 1.2.3.4\n\n")
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		resolvConfPath := file.Name()

		defer func() {
			err := os.Remove(resolvConfPath)
			require.NoError(t, err)
		}()

		ip := net.IP{1, 1, 1, 1}
		const keepNameserver = true

		err = UseDNSSystemWide(resolvConfPath, ip, keepNameserver)

		require.NoError(t, err)

		file, err = os.Open(resolvConfPath)
		require.NoError(t, err)
		b, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\nnameserver 1.2.3.4\n", string(b))
	})
}
