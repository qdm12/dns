package doh

type Metrics interface {
	DoHDialInc(url string)
}
