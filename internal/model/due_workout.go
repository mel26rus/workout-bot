package model

// DueWorkout - тренировка, которую надо отправить человеку сегодня (строка doc_tab_user_workouts)
type DueWorkout struct {
	ID              int64 // doc_tab_user_workouts.id
	TgID            int64
	DocTabWorkoutID int64
	DayNumber       int
	Description     string
	FileName        string
	MimeType        string
	TgFileID        string
	HasFile         bool
}
