package service

import (
	"context"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type DocWorkoutService struct {
	repo *repository.DocWorkoutRepo
}

func NewDocWorkoutService(r *repository.DocWorkoutRepo) *DocWorkoutService {
	return &DocWorkoutService{repo: r}
}

func (s *DocWorkoutService) GetAll(ctx context.Context) ([]model.DocWorkout, error) {
	return s.repo.GetAll(ctx)
}

func (s *DocWorkoutService) GetByID(ctx context.Context, id int64) (model.DocWorkout, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DocWorkoutService) Create(ctx context.Context, w model.DocWorkout) (model.DocWorkout, error) {
	return s.repo.Create(ctx, w)
}
