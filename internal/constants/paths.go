package constants

import (
	"github.com/qdm12/cloudflare-dns-server/internal/models"
)

const (
	// UnboundConf is the file path to the Unbound configuration file
	UnboundConf models.Filepath = "/unbound/unbound.conf"
	// ResolvConf is the file path to the system resolv.conf file
	ResolvConf models.Filepath = "/etc/resolv.conf"
	// CACertificates is the file path to the CA certificates file
	CACertificates models.Filepath = "/unbound/ca-certificates.crt"
	// RootHints is the filepath to the root.hints file used by Unbound
	RootHints models.Filepath = "/unbound/root.hints"
	// RootKey is the filepath to the root.key file used by Unbound
	RootKey models.Filepath = "/unbound/root.key"
)
