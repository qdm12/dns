// Package prometheus offers a New function to return a Prometheus
// HTTP server together with all the metrics registered.
package prometheus

import (
	"fmt"
)

type Erroer interface {
	Error(s string)
}

type promLogger struct {
	logger Erroer
}

func (p *promLogger) Println(v ...interface{}) {
	message := fmt.Sprint(v...)
	p.logger.Error(message)
}
