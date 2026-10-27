package service

import (
	"context"
	"time"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type DocUserWorkoutService struct {
	repo *repository.DocUserWorkoutRepo
	loc  *time.Location
}

func NewDocUserWorkoutService(r *repository.DocUserWorkoutRepo, loc *time.Location) *DocUserWorkoutService {
	return &DocUserWorkoutService{repo: r, loc: loc}
}

// ConfirmPayment подтверждает оплату: старт челленджа на следующий день по времени из конфига (МСК)
func (s *DocUserWorkoutService) ConfirmPayment(ctx context.Context, userID int64) (model.PaymentConfirmation, error) {
	now := time.Now().In(s.loc)
	startDate := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	return s.repo.ConfirmPayment(ctx, userID, startDate)
}
