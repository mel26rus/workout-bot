package dateutil

import "time"

// Today возвращает сегодняшнюю дату в часовом поясе loc (как дату в UTC без времени)
func Today(loc *time.Location) time.Time {
	now := time.Now().In(loc)

	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
