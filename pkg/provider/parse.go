package provider

import (
	"errors"
	"fmt"
	"strings"
)

var ErrParse = errors.New("cannot parse provider")

func Parse(s string) (provider Provider, err error) {
	for _, provider := range All() {
		if strings.EqualFold(s, provider.String()) {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("%w: %q", ErrParse, s)
}
