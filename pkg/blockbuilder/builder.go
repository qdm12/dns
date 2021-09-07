package blockbuilder

import (
	"context"
	"net/http"
)

var _ Interface = (*Builder)(nil)

type Interface interface {
	All(ctx context.Context, settings Settings) Result
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
