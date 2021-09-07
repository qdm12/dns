package log

import "github.com/qdm12/dns/pkg/log/noop"

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Warner

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
