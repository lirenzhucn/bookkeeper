package bookkeeper

func stringInList(s string, l []string) bool {
	for _, ss := range l {
		if s == ss {
			return true
		}
	}
	return false
}
