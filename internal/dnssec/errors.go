package dnssec

import "errors"

var (
	// TODO review exported errors usage and all sentinel errors.
	ErrBogus = errors.New("bogus response")
)

var _ error = (*joinedErrors)(nil)

type joinedErrors struct { //nolint:errname
	errs []error
}

func (e *joinedErrors) add(err error) {
	e.errs = append(e.errs, err)
}

func (e *joinedErrors) Error() string {
	return joinStrings(e.errs, "and")
}

func (e *joinedErrors) Unwrap() []error {
	return e.errs
}
