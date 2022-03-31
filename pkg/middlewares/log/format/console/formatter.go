package console

type Formatter struct{} //nolint:errname

func New(settings Settings) *Formatter {
	return &Formatter{}
}
