package localdns

type LocalChecker interface {
	IsFQDNLocal(fqdn string) bool
}

type Logger interface {
	Debug(message string)
	Warn(message string)
}
