package handler

import "time"

// unixToTime 将 Unix 时间戳(秒)转 time.Time
func unixToTime(unix int64) time.Time {
	if unix <= 0 {
		return time.Now()
	}
	return time.Unix(unix, 0)
}
