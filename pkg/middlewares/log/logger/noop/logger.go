package noop

import "github.com/miekg/dns"

type Logger struct{}

func New() *Logger {
	return new(Logger)
}

func (l *Logger) Error(id uint16, errMessage string) {}
func (l *Logger) Log(request, response *dns.Msg)     {}
