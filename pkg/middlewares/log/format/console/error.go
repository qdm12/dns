package console

import "fmt"

func (f *Formatter) Error(requestID uint16, errString string) string {
	return "id: " + fmt.Sprint(requestID) + "; error: " + errString
}
