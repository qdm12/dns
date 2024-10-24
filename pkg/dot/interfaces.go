package dot

type Metrics interface {
	DoTDialInc(provider, address, outcome string)
}

type Warner interface {
	Warn(s string)
}
