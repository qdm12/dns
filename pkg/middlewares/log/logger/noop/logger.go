package noop

type Logger struct{}

func New() *Logger {
	return new(Logger)
}

func (l *Logger) Error(s string)              {}
func (l *Logger) LogRequest(s string)         {}
func (l *Logger) LogResponse(s string)        {}
func (l *Logger) LogRequestResponse(s string) {}
