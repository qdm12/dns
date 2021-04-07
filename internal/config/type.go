package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qdm12/golibs/params"
)

type UpstreamType string

const (
	DoT UpstreamType = "DoT"
	DoH UpstreamType = "DoH"
)

var ErrInvalidUpstreamType = errors.New("invalid upstream type")

func getUpstreamType(env params.Env) (ut UpstreamType, err error) {
	s, err := env.Get("UPSTREAM_TYPE", params.Default(string(DoT)))
	if err != nil {
		return "", fmt.Errorf("environment variable UPSTREAM_TYPE: %w", err)
	}
	switch s {
	case strings.ToLower(string(DoT)):
		return DoT, nil
	case strings.ToLower(string(DoH)):
		return DoH, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidUpstreamType, s)
	}
}
