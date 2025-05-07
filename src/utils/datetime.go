package utils

import "time"

func GetUTCZeroOfNDaysLater(n int) time.Time {
	now := time.Now().UTC()
	target := now.AddDate(0, 0, n)
	return time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
}
