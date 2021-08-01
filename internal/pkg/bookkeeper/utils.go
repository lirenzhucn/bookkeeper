package bookkeeper

import "regexp"

func stringInList(s string, l []string) bool {
	for _, ss := range l {
		if s == ss {
			return true
		}
	}
	return false
}

func MaskDbPassword(msg string) string {
	r, err := regexp.Compile("(postgres://.+:)(.+)@")
	if err != nil {
		return msg
	}
	return r.ReplaceAllString(msg, "${1}[REDACTED]@")
}
