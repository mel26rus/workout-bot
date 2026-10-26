package model

// DocTabWorkout - тренировка челленджа (таблица doc_tab_workouts)
type DocTabWorkout struct {
	ID           int64
	DocWorkoutID int64
	DayNumber    int
	Description  string
	File         []byte
	FileName     string
	MimeType     string
}

// DocTabWorkoutInfo - строка списка тренировок, без самого файла
type DocTabWorkoutInfo struct {
	ID             int64
	DocWorkoutID   int64
	DocWorkoutName string
	DayNumber      int
	Description    string
	FileName       string
	FileSize       int64
}
