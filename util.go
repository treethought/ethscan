package main

import (
	"time"
)

func formatTime(t time.Time) string {
	return t.Format("01/02/06 3:04 pm")
}

func formatUnixTime(t uint64) string {
	tm := time.Unix(int64(t), 0)
	return formatTime(tm)
}

