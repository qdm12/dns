package nameserver

import (
	"context"
	"io/ioutil"
	"net"
	"strings"

	"github.com/qdm12/golibs/os"
)

// UseDNSInternally is to change the Go program DNS only.
func UseDNSInternally(ip net.IP) { //nolint:interfacer
	net.DefaultResolver.PreferGo = true
	net.DefaultResolver.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
	}
}

const resolvConfFilepath = "/etc/resolv.conf"

// UseDNSSystemWide changes the nameserver to use for DNS system wide.
func UseDNSSystemWide(openFile os.OpenFileFunc, ip net.IP, keepNameserver bool) error { //nolint:interfacer
	file, err := openFile(resolvConfFilepath, os.O_RDONLY, 0)
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

	file, err = openFile(resolvConfFilepath, os.O_WRONLY|os.O_TRUNC, 0644)
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
