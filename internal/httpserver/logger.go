package httpserver

type Infoer interface {
	Info(message string)
}

type noopLogger struct{}

func (noopLogger) Info(_ string) {}
