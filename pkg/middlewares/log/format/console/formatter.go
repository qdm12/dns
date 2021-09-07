package console

type Formatter struct {
	// Cache string serializations
	idToRequestString  map[uint16]string
	idToResponseString map[uint16]string
}

func New() *Formatter {
	return &Formatter{
		idToRequestString:  make(map[uint16]string),
		idToResponseString: make(map[uint16]string),
	}
}
