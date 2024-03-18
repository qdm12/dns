package dnssec

import "fmt"

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

	result = fmt.Sprint(elements[0])
	for i := 1; i < len(elements); i++ {
		lastElement := i == len(elements)-1
		if lastElement {
			result += " " + lastJoin + " " + fmt.Sprint(elements[i])
			continue
		}
		result += ", " + fmt.Sprint(elements[i])
	}

	return result
}
