package dateutils

import (
	"time"
)

// SQLiteDateForm is the date format for SQLite
const SQLiteDateForm = "2006-01-02 15:04:05"

// GetDateTimeNowMilliSecond returns the current time down to millisecond
func GetDateTimeNowMilliSecond() time.Time {
	creationTime := time.Now()
	return time.Date(
		creationTime.Year(),
		creationTime.Month(),
		creationTime.Day(),
		creationTime.Hour(),
		creationTime.Minute(),
		creationTime.Second(),
		0,
		time.UTC,
	)
}
