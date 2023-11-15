package doh

func ptrTo[T any](value T) *T {
	return &value
}
