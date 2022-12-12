// Package prometheus offers a New function to return a Prometheus
// HTTP server together with all the metrics registered.
package prometheus

import (
	"fmt"
)

type Logger interface {
	Info(s string)
	Error(s string)
}

type promLogger struct {
	logger Logger
}

func (p *promLogger) Println(v ...interface{}) {
	message := fmt.Sprint(v...)
	p.logger.Error(message)
}
