package model

import "time"

// Participant - строка списка занимающихся: человек и его челлендж
type Participant struct {
	UserID         int64
	TgID           int64
	Username       string
	FirstName      string
	Surname        string
	DocWorkoutName string
	IsPaid         bool
	StartDate      *time.Time
	Total          int64
	Sent           int64
	Done           int64
}
