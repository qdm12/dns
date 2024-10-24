package server

import (
	"context"

	"github.com/miekg/dns"
)

type exchangerIntf interface {
	Exchange(ctx context.Context, network string, request *dns.Msg) (
		response *dns.Msg, err error,
	)
}
