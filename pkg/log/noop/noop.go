package noop

type Logger struct{}

func New() *Logger {
	return new(Logger)
}

func (l *Logger) Debug(s string) {}
func (l *Logger) Info(s string)  {}
func (l *Logger) Warn(s string)  {}
func (l *Logger) Error(s string) {}
