package dnssec

type Logger interface {
	Info(message string)
	Warn(message string)
}
