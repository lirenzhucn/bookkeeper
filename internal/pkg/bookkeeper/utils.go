package bookkeeper

import (
	"log"
	"regexp"

	"go.uber.org/zap"
)

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
