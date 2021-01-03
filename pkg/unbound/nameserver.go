package unbound

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

// UseDNSInternally is to change the Go program DNS only.
func (c *configurator) UseDNSInternally(ip net.IP) {
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
	}
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
func (c *configurator) UseDNSSystemWide(ip net.IP, keepNameserver bool) error {
	file, err := c.openFile(resolvConfFilepath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		_ = file.Close()
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
	_, err = file.WriteString(s)
	if err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
