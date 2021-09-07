package logger

import "github.com/qdm12/dns/pkg/middlewares/log/logger/noop"

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Interface

var _ Interface = (*noop.Logger)(nil)

type Interface interface {
	// Error logs errors returned by the DNS handler using
	// the formatter Error method.
	Error(s string)
	// LogRequest logs the request using the formatter Request
	// method, at the beginning of the request handling.
	LogRequest(s string)
	// LogResponse logs the response using the formatter Response
	// method, at the end of the request handling.
	LogResponse(s string)
	// LogRequestResponse logs the request and response together
	// using the formatter RequestResponse method, at the end of the
	// request handling.
	LogRequestResponse(s string)
}
