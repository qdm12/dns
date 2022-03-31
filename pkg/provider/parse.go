package provider

import (
	"errors"
	"fmt"
	"strings"
)

var ErrParse = errors.New("provider does not match any known providers")

func Parse(s string) (provider Provider, err error) {
	for _, provider := range All() {
		if strings.EqualFold(s, provider.Name) {
			return provider, nil
		}
	}
	return provider, fmt.Errorf("%w: %s", ErrParse, s)
}
