package blockbuilder

import (
	"context"
	"net/http"

	"inet.af/netaddr"
)

var _ Interface = (*Builder)(nil)

type Interface interface {
	All(ctx context.Context, settings Settings) (
		blockedHostnames []string, blockedIPs []netaddr.IP,
		blockedIPPrefixes []netaddr.IPPrefix, errs []error)
}

func New(client *http.Client) *Builder {
	return &Builder{
		client: client,
		// TODO cache blocked IPs and hostnames after first request?
	}
}

type Builder struct {
	client *http.Client
}
