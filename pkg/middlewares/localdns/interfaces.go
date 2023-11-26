package localdns

type Logger interface {
	Debug(message string)
	Warn(message string)
}
