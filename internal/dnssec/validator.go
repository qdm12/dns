package dnssec

import "sync/atomic"

type Validator struct {
	rootTrustAnchors atomic.Pointer[trustAnchorSet]
}

func New() *Validator {
	validator := new(Validator)
	validator.setRootTrustAnchors(defaultRootTrustAnchors())
	return validator
}
