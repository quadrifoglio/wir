package utils

// SliceContainsStr returns true if 's'
// is contained within 'arr'
func SliceContainsStr(s string, arr []string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}

	return false
}
