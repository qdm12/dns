package dnssec

import "github.com/qdm12/log"

//nolint:gochecknoglobals
var globalDebugLogger = log.New(log.SetCallerFile(true),
	log.SetCallerLine(true), log.SetComponent("dnssec-debug"), log.SetLevel(log.LevelDebug))
