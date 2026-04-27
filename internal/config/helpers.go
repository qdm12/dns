package config

import "strings"

func andStrings(strings []string) (result string) {
	return joinStrings(strings, "and")
}

func joinStrings(ss []string, lastJoin string) (result string) {
	if len(ss) == 0 {
		return ""
	}

	stringsBuilder := strings.Builder{}
	stringsBuilder.WriteString(ss[0])
	for i := 1; i < len(ss); i++ {
		if i < len(ss)-1 {
			stringsBuilder.WriteString(ss[i] + ", ")
		} else {
			stringsBuilder.WriteString(lastJoin + " " + ss[i])
		}
	}

	return stringsBuilder.String()
}
