package dnssec

import (
	"sync/atomic"

	"github.com/qdm12/dns/v2/internal/local"
)

type Validator struct {
	localChecker     *local.Checker
	rootTrustAnchors atomic.Pointer[trustAnchorSet]
}

func New() *Validator {
	validator := &Validator{
		localChecker: local.New(nil),
	}
	validator.setRootTrustAnchors(defaultRootTrustAnchors())
	return validator
}
