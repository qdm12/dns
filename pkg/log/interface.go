package log

import "github.com/qdm12/dns/pkg/log/noop"

var (
	_ Logger = (*noop.Logger)(nil)
	_ Warner = (*noop.Logger)(nil)
)

type Logger interface {
	Debug(s string)
	Info(s string)
	Warner
	Error(s string)
}

type Warner interface {
	Warn(s string)
}
