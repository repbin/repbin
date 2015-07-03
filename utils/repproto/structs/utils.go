package structs

// BoolToString converts a boolean into "true" or "false"
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// StringToBool converts "true"/"false" into boolean
func StringToBool(s string) bool {
	if s == "true" {
		return true
	}
	return false
}
