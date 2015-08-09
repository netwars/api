package main

import "time"

const (
	timeTodayPrefix     = "Dzisiaj, "
	timeYesterdayPrefix = "Wczoraj, "
	timeTodayLayout     = "Dzisiaj, 15:04:05"
	timeYesterdayLayout = "Wczoraj, 15:04:05"
	timeShortLayout     = "2006-01-02 15:04:05"
	timeGeneralLayout   = "2006-01-02 15:04:05"
)

func parseDate(raw string) (*time.Time, error) {
	var err error
	var t time.Time

	t, err = time.Parse(timeGeneralLayout, raw)
	if err != nil {
		t, err = time.Parse(timeYesterdayLayout, raw)
		if err != nil {
			t, err = time.Parse(timeTodayLayout, raw)
			if err == nil {
				year, month, day := time.Now().Date()
				t = t.AddDate(year, int(month)-1, day-1)
			}
		} else {
			year, month, day := time.Now().Add(-24 * time.Hour).Date()
			t = t.AddDate(year, int(month)-1, day-1)
		}
	}

	return &t, err
}
