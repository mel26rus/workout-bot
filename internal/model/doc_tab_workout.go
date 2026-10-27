package model

// DocTabWorkout - тренировка челленджа (таблица doc_tab_workouts): пост канала, который уходит в этот день
type DocTabWorkout struct {
	DocWorkoutID  int64
	DayNumber     int
	ChannelPostID int64
}
