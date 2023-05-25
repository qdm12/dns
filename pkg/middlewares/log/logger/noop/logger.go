package noop

import (
	"net"

	"github.com/miekg/dns"
)

type Logger struct{}

func New() *Logger {
	return new(Logger)
}

func (l *Logger) Error(uint16, string) {}
func (l *Logger) Log(net.Addr, *dns.Msg, *dns.Msg) {
}
