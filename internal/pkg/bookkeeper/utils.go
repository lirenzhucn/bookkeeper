package bookkeeper

import (
	"log"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

func stringMatchList(s string, l []string) bool {
	for _, ss := range l {
		if stringMatch(s, ss) {
			return true
		}
	}
	return false
}

func stringMatch(s string, m string) bool {
	for _, mm := range strings.Split(m, "|") {
		if strings.HasPrefix(s, mm) {
			return true
		}
	}
	return false
}

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

func SetupZapGlobals() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)
}
