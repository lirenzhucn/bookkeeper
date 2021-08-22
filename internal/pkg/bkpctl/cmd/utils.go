package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type dateRange struct {
	startDate time.Time
	endDate   time.Time
}

var numDaysMatcher = regexp.MustCompile(`^past\s*(\d+)\s*day(s*)$`)

func parseDateRange(dateRangeStr string) (dr dateRange, ok bool) {
	matchedGroups := numDaysMatcher.FindStringSubmatch(dateRangeStr)
	switch {
	case dateRangeStr == "past week":
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * 6)
		dr = dateRange{
			startDate: cutoff,
			endDate:   today,
		}
		ok = true
	case dateRangeStr == "last week":
		today := time.Now().Add(-time.Hour * 24 * 7)
		cutoff := today.Add(-time.Hour * 24 * 6)
		dr = dateRange{
			startDate: cutoff,
			endDate:   today,
		}
		ok = true
	case matchedGroups != nil:
		numDays, _ := strconv.ParseInt(matchedGroups[1], 10, 64)
		today := time.Now()
		cutoff := today.Add(-time.Hour * 24 * time.Duration(numDays-1))
		dr = dateRange{
			startDate: cutoff,
			endDate:   today,
		}
		ok = true
	default:
		ok = false
	}
	return
}

func dateRangeToQuery(dateRangeStr string) (parsed string, ok bool) {
	var dr dateRange
	if dr, ok = parseDateRange(dateRangeStr); !ok {
		return
	}
	parsed = fmt.Sprintf(
		"date>=%s AND date<=%s", dr.startDate.Format("2006/01/02"),
		dr.endDate.Format("2006/01/02"),
	)
	return
}
