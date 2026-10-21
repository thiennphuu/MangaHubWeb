package utils

import (
	"time"
)

// TimeHelper provides time utility functions
type TimeHelper struct{}

// NewTimeHelper creates a new time helper
func NewTimeHelper() *TimeHelper {
	return &TimeHelper{}
}

// Now returns the current time in UTC
func (th *TimeHelper) Now() time.Time {
	return time.Now().UTC()
}

// NowUnix returns the current Unix timestamp
func (th *TimeHelper) NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli returns the current Unix timestamp in milliseconds
func (th *TimeHelper) NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// FormatTime formats a time to string
func (th *TimeHelper) FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses a string to time
func (th *TimeHelper) ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// AddHours adds hours to a time
func (th *TimeHelper) AddHours(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

// AddDays adds days to a time
func (th *TimeHelper) AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}
