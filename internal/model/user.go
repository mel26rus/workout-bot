package model

// User - пользователь Telegram (таблица users)
type User struct {
	ID           int64
	TgID         int64
	Username     string
	FirstName    string
	Surname      string
	LanguageCode string
}
