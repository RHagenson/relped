package util

func StringInSlice(str string, list []string) bool {
	for _, elm := range list {
		if elm == str {
			return true
		}
	}
	return false
}