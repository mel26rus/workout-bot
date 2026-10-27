package model

import "time"

// PaymentConfirmation - результат подтверждения оплаты, из него собирается сообщение человеку
type PaymentConfirmation struct {
	TgID           int64
	DocWorkoutName string
	StartDate      time.Time
}
