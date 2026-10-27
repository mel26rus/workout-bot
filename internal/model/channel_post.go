package model

import "time"

// ChannelPost - пост закрытого канала тренера (таблица channel_posts)
type ChannelPost struct {
	ID        int64
	ChatID    int64
	MessageID int
	Text      string
	MediaType string
	FileName  string
	FileSize  int64
	PostDate  time.Time
}

// ChannelPostInfo - строка таблицы на странице тренировок: пост и день, к которому он привязан
type ChannelPostInfo struct {
	ChannelPost
	DocWorkoutID int64 // 0: пост ещё не привязан
	DayNumber    int   // 0: пост ещё не привязан
}
