package settings

func andStrings(strings []string) (result string) {
	if len(strings) == 0 {
		return ""
	}

	result = strings[0]
	for i := 1; i < len(strings); i++ {
		if i < len(strings)-1 {
			result += strings[i] + ", "
		} else {
			result += " and " + strings[i]
		}
	}

	return result
}

func boolToEnabled(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
