package dnssec

import (
	"fmt"
	"strings"
)

func quoteStrings(elements []string) {
	for i := range elements {
		elements[i] = "\"" + elements[i] + "\""
	}
}

func orStrings[T comparable](elements []T) (result string) {
	return joinStrings(elements, "or")
}

func joinStrings[T comparable](elements []T, lastJoin string) (result string) {
	if len(elements) == 0 {
		return ""
	}

	builder := strings.Builder{}
	fmt.Fprint(&builder, elements[0])
	for i := 1; i < len(elements); i++ {
		lastElement := i == len(elements)-1
		if lastElement {
			fmt.Fprintf(&builder, " "+lastJoin+" %v", elements[i])
			continue
		}
		fmt.Fprintf(&builder, ", %v", elements[i])
	}

	return builder.String()
}
