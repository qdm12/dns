package nameserver

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// UseDNSInternally is to change the Go program DNS only.
func UseDNSInternally(ip net.IP) { //nolint:interfacer
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
	}
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
// If resolvConfPath is empty, it defaults to /etc/resolv.conf.
func UseDNSSystemWide(resolvConfPath string, ip net.IP, keepNameserver bool) error { //nolint:interfacer
	const defaultResolvConfPath = "/etc/resolv.conf"
	if resolvConfPath == "" {
		resolvConfPath = defaultResolvConfPath
	}
	file, err := os.Open(resolvConfPath)
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
		"nameserver " + ip.String(),
	}
	for _, line := range strings.Split(s, "\n") {
		if line == "" ||
			(!keepNameserver && strings.HasPrefix(line, "nameserver ")) {
			continue
		}
		lines = append(lines, line)
	}

	s = strings.Join(lines, "\n") + "\n"

	file, err = os.OpenFile(resolvConfPath, os.O_WRONLY|os.O_TRUNC, 0644)
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
