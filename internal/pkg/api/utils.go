package api

import (
	"fmt"
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

type dateRange struct {
	startDate time.Time
	endDate   time.Time
}

func parseMultipleDateRangesInQueryAndFail(
	w http.ResponseWriter, r *http.Request, queryTerm string,
) (dateRanges []dateRange, ok bool) {
	ok = false
	dateRangeStr := r.FormValue(queryTerm)
	if dateRangeStr == "" {
		http.Error(w, "Invalid query string", 400)
		return
	}
	offset, _ := time.ParseDuration("23h59m59s")
	for _, s := range strings.Split(dateRangeStr, ",") {
		p := strings.Split(s, "-")
		if len(p) != 2 {
			http.Error(w, "Invalid date range", 400)
			return
		}
		startDate, err := time.Parse("2006/01/02", p[0])
		if err != nil {
			http.Error(w, "Invalid date", 400)
			return
		}
		endDate, err := time.Parse("2006/01/02", p[1])
		if err != nil {
			http.Error(w, "Invalid date", 400)
			return
		}
		// shift endDate to the end of that day
		endDate = endDate.Add(offset)
		dateRanges = append(dateRanges,
			dateRange{startDate: startDate, endDate: endDate})
	}
	ok = true
	return
}

func parseMultipleDateTimesInQueryAndFail(
	w http.ResponseWriter, r *http.Request, queryTerm string,
) (dates []time.Time, ok bool) {
	ok = true
	dateStr := r.FormValue(queryTerm)
	if dateStr == "" {
		http.Error(w, "Invalid query string", 400)
		ok = false
		return
	}
	offset, _ := time.ParseDuration("23h59m59s")
	for _, s := range strings.Split(dateStr, ",") {
		d, err := time.Parse("2006/01/02", s)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid query term %s", queryTerm), 400)
			ok = false
			return
		}
		d = d.Add(offset)
		dates = append(dates, d)
	}
	return
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
