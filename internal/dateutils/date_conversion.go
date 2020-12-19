package dateutils

import (
	"time"
)

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
