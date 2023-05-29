package nameserver

import (
	"io/fs"
	"net/netip"
	"os"
	"strings"

	"github.com/qdm12/gosettings"
)

type SettingsSystemDNS struct {
	// IP is the IP address to use for the DNS.
	// It defaults to 127.0.0.1 if nil.
	IP netip.Addr
	// ResolvPath is the path to the resolv configuration file.
	// It defaults to /etc/resolv.conf.
	ResolvPath string
	// KeepNameserver can be set to preserve existing nameserver lines
	// in the resolv configuration file.
	KeepNameserver bool
}

func (s *SettingsSystemDNS) SetDefaults() {
	s.IP = gosettings.DefaultValidator(s.IP, netip.AddrFrom4([4]byte{127, 0, 0, 1}))
	s.ResolvPath = gosettings.DefaultString(s.ResolvPath, "/etc/resolv.conf")
}

func (s *SettingsSystemDNS) Validate() (err error) {
	// TODO check s.ResolvPath file exists
	return nil
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
// If resolvConfPath is empty, it defaults to /etc/resolv.conf.
func UseDNSSystemWide(settings SettingsSystemDNS) (err error) {
	settings.SetDefaults()

	data, err := os.ReadFile(settings.ResolvPath)
	if err != nil {
		return err
	}

	s := strings.TrimSuffix(string(data), "\n")

	lines := []string{
		"nameserver " + settings.IP.String(),
	}
	for _, line := range strings.Split(s, "\n") {
		if line == "" ||
			(!settings.KeepNameserver && strings.HasPrefix(line, "nameserver ")) {
			continue
		}
		lines = append(lines, line)
	}

	s = strings.Join(lines, "\n") + "\n"

	const permissions fs.FileMode = 0600
	return os.WriteFile(settings.ResolvPath, []byte(s), permissions)
}
