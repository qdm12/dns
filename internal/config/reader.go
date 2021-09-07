package config

import (
	"github.com/qdm12/golibs/logging"
	"github.com/qdm12/golibs/params"
	"github.com/qdm12/golibs/verification"
)

var _ SettingsReader = (*Reader)(nil)

type SettingsReader interface {
	ReadSettings() (s Settings, err error)
}

type Reader struct {
	env      params.Interface
	logger   logging.Logger
	verifier verification.Verifier
}

func NewReader(logger logging.Logger) *Reader {
	return &Reader{
		env:      params.New(),
		logger:   logger,
		verifier: verification.NewVerifier(),
	}
}

func (r *Reader) ReadSettings() (s Settings, err error) {
	err = s.get(r)
	return s, err
}
