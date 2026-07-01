package utils

import "time"

func Now() time.Time {
	return time.Now()
}

func NowUnix() int64 {
	return time.Now().Unix()
}

func TimestampToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func GetYearFromTimestamp(ts int64) int {
	return time.Unix(ts, 0).Year()
}
