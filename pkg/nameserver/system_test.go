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

		settings := SettingsSystemDNS{
			ResolvPath: filepath.Join(dirPath, "resolv.conf"),
			IP:         net.IP{1, 1, 1, 1},
		}

		err = UseDNSSystemWide(settings)

		require.Error(t, err)
		assert.Equal(t, "open "+settings.ResolvPath+": no such file or directory", err.Error())
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

		settings := SettingsSystemDNS{
			ResolvPath: resolvConfPath,
			IP:         net.IP{1, 1, 1, 1},
		}

		err = UseDNSSystemWide(settings)

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

		settings := SettingsSystemDNS{
			ResolvPath:     resolvConfPath,
			IP:             net.IP{1, 1, 1, 1},
			KeepNameserver: true,
		}

		err = UseDNSSystemWide(settings)

		require.NoError(t, err)

		file, err = os.Open(resolvConfPath)
		require.NoError(t, err)
		b, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\nnameserver 1.2.3.4\n", string(b))
	})
}
