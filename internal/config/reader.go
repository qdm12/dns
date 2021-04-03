package config

import (
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/golibs/verification"
)

//go:generate mockgen -destination=mock_$GOPACKAGE/$GOFILE . Reader

type Reader interface {
	ReadSettings() (s Settings, err error)
}

type reader struct {
	env      params.Env
	logger   logging.Logger
	verifier verification.Verifier
}

func NewReader(logger logging.Logger) Reader {
	return &reader{
		env:      params.NewEnv(),
		logger:   logger,
		verifier: verification.NewVerifier(),
	}
}

func (r *reader) ReadSettings() (s Settings, err error) {
	err = s.get(r)
	return s, err
}

func (r *reader) onRetroActive(oldKey, newKey string) {
	r.logger.Warn(
		"You are using the old environment variable %s, please consider changing it to %s",
		oldKey, newKey,
	)
}
