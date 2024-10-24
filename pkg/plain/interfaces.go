package plain

type Metrics interface {
	PlainDialInc(address, outcome string)
}

type Warner interface {
	Warn(s string)
}
