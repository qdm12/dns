package nameserver

import (
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrTo[T any](value T) *T {
	return &value
}

func Test_UseDNSSystemWide(t *testing.T) {
	t.Parallel()

	t.Run("file_does_not_exist", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		resolvConfPath := filepath.Join(dirPath, "resolv.conf")

		settings := SettingsSystemDNS{
			ResolvPath: resolvConfPath,
			IP:         netip.AddrFrom4([4]byte{1, 1, 1, 1}),
		}

		err := UseDNSSystemWide(settings)

		require.NoError(t, err)
		data, err := os.ReadFile(resolvConfPath)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\n", string(data))
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		resolvConfPath := filepath.Join(dirPath, "resolv.conf")

		file, err := os.Create(resolvConfPath)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		settings := SettingsSystemDNS{
			ResolvPath: resolvConfPath,
			IP:         netip.AddrFrom4([4]byte{1, 1, 1, 1}),
		}

		err = UseDNSSystemWide(settings)

		require.NoError(t, err)

		require.NoError(t, err)
		data, err := os.ReadFile(settings.ResolvPath)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\n", string(data))
	})

	t.Run("preserve nameserver", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		resolvConfPath := filepath.Join(dirPath, "resolv.conf")

		file, err := os.Create(resolvConfPath)
		require.NoError(t, err)
		_, err = io.WriteString(file, "nameserver 1.2.3.4\n\n")
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		settings := SettingsSystemDNS{
			ResolvPath:     resolvConfPath,
			IP:             netip.AddrFrom4([4]byte{1, 1, 1, 1}),
			KeepNameserver: ptrTo(true),
		}

		err = UseDNSSystemWide(settings)

		require.NoError(t, err)

		require.NoError(t, err)
		data, err := os.ReadFile(settings.ResolvPath)
		require.NoError(t, err)
		assert.Equal(t, "nameserver 1.1.1.1\nnameserver 1.2.3.4\n", string(data))
	})
}
