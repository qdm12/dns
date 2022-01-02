package nameserver

import (
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/qdm12/dns/internal/settings/defaults"
)

type SettingsSystemDNS struct {
	// IP is the IP address to use for the DNS.
	// It defaults to 127.0.0.1 if nil.
	IP net.IP
	// ResolvPath is the path to the resolv configuration file.
	// It defaults to /etc/resolv.conf.
	ResolvPath string
	// KeepNameserver can be set to preserve existing nameserver lines
	// in the resolv configuration file.
	KeepNameserver bool
}

func (s *SettingsSystemDNS) SetDefaults() {
	s.IP = defaults.IP(s.IP, net.IPv4(127, 0, 0, 1)) //nolint:gomnd
	s.ResolvPath = defaults.String(s.ResolvPath, "/etc/resolv.conf")
}

func (s *SettingsSystemDNS) Validate() (err error) {
	// TODO check s.ResolvPath file exists
	return nil
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
// If resolvConfPath is empty, it defaults to /etc/resolv.conf.
func UseDNSSystemWide(settings SettingsSystemDNS) (err error) {
	settings.SetDefaults()

	file, err := os.Open(settings.ResolvPath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
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

	file, err = os.OpenFile(settings.ResolvPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = file.WriteString(s)
	if err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
