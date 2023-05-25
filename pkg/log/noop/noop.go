package noop

type Logger struct{}

func New() *Logger {
	return new(Logger)
}

func (l *Logger) Debug(_ string) {}
func (l *Logger) Info(_ string)  {}
func (l *Logger) Warn(_ string)  {}
func (l *Logger) Error(_ string) {}
