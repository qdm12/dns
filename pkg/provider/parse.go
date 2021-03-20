package provider

import (
	"errors"
	"fmt"
)

var ErrParse = errors.New("cannot parse provider")

func Parse(s string) (provider Provider, err error) {
	switch s {
	case "cirafamily":
		provider = CiraFamily()
	case "ciraprivate":
		provider = CiraPrivate()
	case "ciraprotected":
		provider = CiraProtected()
	case "cleanbrowsingadult":
		provider = CleanBrowsingAdult()
	case "cleanbrowsingfamily":
		provider = CleanBrowsingFamily()
	case "cleanbrowsingsecurity":
		provider = CleanBrowsingSecurity()
	case "cloudflare":
		provider = Cloudflare()
	case "cloudflarefamily":
		provider = CloudflareFamily()
	case "cloudflaresecurity":
		provider = CloudflareSecurity()
	case "google":
		provider = Google()
	case "libredns":
		provider = LibreDNS()
	case "quad9":
		provider = Quad9()
	case "quad9secured":
		provider = Quad9Secured()
	case "quad9unsecured":
		provider = Quad9Unsecured()
	case "quadrant":
		provider = Quadrant()
	default:
		return nil, fmt.Errorf("%w: %q", ErrParse, s)
	}
	return provider, nil
}
