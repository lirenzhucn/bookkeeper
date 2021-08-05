package api

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

func checkErr(err error, w http.ResponseWriter, statusCode int,
	msg string, a ...interface{}) bool {
	sugar := zap.L().Sugar()
	defer sugar.Sync()

	if err != nil {
		a = append(a, "error")
		a = append(a, err)
		sugar.Errorw(msg, a...)
		var outMsg = msg
		if statusCode == 500 {
			outMsg = "Internal Server Error"
		}
		http.Error(w, outMsg, statusCode)
		return false
	}
	return true
}

func parseDateTimeInQueryAndFail(
	w http.ResponseWriter, r *http.Request, queryTerm string,
) (date time.Time, ok bool) {
	date = time.Now()
	ok = true
	dateStr := r.FormValue(queryTerm)
	if dateStr == "" {
		http.Error(w, "Invalid query string", 400)
		ok = false
		return
	}
	date, err := time.Parse("2006/01/02", dateStr)
	if err != nil {
		http.Error(w, "Invalid query term date", 400)
		ok = false
		return
	}
	offset, _ := time.ParseDuration("23h59m59s")
	date = date.Add(offset)
	return
}

func parseTagsInQueryAndFail(
	w http.ResponseWriter, r *http.Request, queryTerm string,
) (tags []string, ok bool) {
	ok = true
	tagsStr := r.FormValue(queryTerm)
	if tagsStr == "" {
		http.Error(w, "Invalid query string", 400)
		ok = false
		return
	}
	tags = strings.Split(tagsStr, ",")
	return
}
