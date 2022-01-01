package provider

import (
	"errors"
	"fmt"
	"strings"
)

var ErrParse = errors.New("provider does not match any known providers")

func Parse(s string) (provider Provider, err error) {
	for _, provider := range All() {
		if strings.EqualFold(s, provider.String()) {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrParse, s)
}
