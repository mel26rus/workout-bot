package service

import (
	"context"
	"time"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type DocTabUserWorkoutService struct {
	repo *repository.DocTabUserWorkoutRepo
}

func NewDocTabUserWorkoutService(r *repository.DocTabUserWorkoutRepo) *DocTabUserWorkoutService {
	return &DocTabUserWorkoutService{repo: r}
}

func (s *DocTabUserWorkoutService) GetDue(ctx context.Context, date time.Time) ([]model.DueWorkout, error) {
	return s.repo.GetDue(ctx, date)
}

func (s *DocTabUserWorkoutService) MarkSent(ctx context.Context, id int64) error {
	return s.repo.MarkSent(ctx, id)
}

func (s *DocTabUserWorkoutService) MarkDone(ctx context.Context, id int64, tgID int64) error {
	return s.repo.MarkDone(ctx, id, tgID)
}
