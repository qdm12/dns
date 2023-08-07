package dns

import (
	"context"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/blockbuilder"
)

type Service interface {
	String() string
	Start() (runError <-chan error, startErr error)
	Stop() error
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Cache interface {
	Add(request, response *dns.Msg)
	Get(request *dns.Msg) (response *dns.Msg)
	Remove(request *dns.Msg)
}

type BlockBuilder interface {
	BuildAll(ctx context.Context) blockbuilder.Result
}
