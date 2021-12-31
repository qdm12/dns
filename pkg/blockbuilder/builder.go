package blockbuilder

import (
	"context"
	"net/http"
)

var _ Interface = (*Builder)(nil)

type Interface interface {
	BuildAll(ctx context.Context, settings BuildSettings) Result
}

func New(settings Settings) *Builder {
	settings.SetDefaults()

	return &Builder{
		client: settings.Client,
		// TODO cache blocked IPs and hostnames after first request?
	}
}

type Builder struct {
	client *http.Client
}
