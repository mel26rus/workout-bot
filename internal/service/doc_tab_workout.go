package service

import (
	"context"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type DocTabWorkoutService struct {
	repo *repository.DocTabWorkoutRepo
}

func NewDocTabWorkoutService(r *repository.DocTabWorkoutRepo) *DocTabWorkoutService {
	return &DocTabWorkoutService{repo: r}
}

// SaveAll возвращает, сколько дней добавлено и сколько заменено
func (s *DocTabWorkoutService) SaveAll(ctx context.Context, list []model.DocTabWorkout) (int, int, error) {
	return s.repo.SaveAll(ctx, list)
}
