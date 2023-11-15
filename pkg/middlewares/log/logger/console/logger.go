package console

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/miekg/dns"
)

// Logger is a DNS middleware logger for consoles.
type Logger struct {
	writer      io.Writer
	logRequest  bool
	logResponse bool
	timeNow     func() time.Time
}

// New creates a new console middleware logger.
func New(settings Settings) (logger *Logger, err error) {
	settings.SetDefaults()

	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("settings validation: %w", err)
	}

	return &Logger{
		writer:      settings.Writer,
		logRequest:  *settings.LogRequests,
		logResponse: *settings.LogResponses,
		timeNow:     time.Now,
	}, nil
}

func (l *Logger) Error(id uint16, errMessage string) {
	l.write(formatError(id, errMessage))
}

func (l *Logger) Log(remoteAddr net.Addr,
	request, response *dns.Msg) {
	var message string
	switch {
	case !l.logRequest && !l.logResponse:
		return
	case l.logRequest && l.logResponse:
		message = formatRequestResponse(request, response)
	case l.logRequest:
		message = formatRequest(request)
	case l.logResponse:
		message = formatResponse(response)
	}
	message = fmt.Sprintf("%s %s %s\n",
		l.timeNow().Format(time.RFC3339),
		remoteAddr, message)

	l.write(message)
}

func (l *Logger) write(message string) {
	_, err := l.writer.Write([]byte(message))
	if err != nil {
		panic(fmt.Sprintf("failed to write to log: %s", err))
	}
}
