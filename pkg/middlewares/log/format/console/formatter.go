package console

type Formatter struct{}

func New(settings Settings) *Formatter {
	return &Formatter{}
}
