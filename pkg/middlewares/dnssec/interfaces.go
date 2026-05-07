package dnssec

type Logger interface {
	Debug(message string)
	Info(message string)
	Warn(message string)
}
