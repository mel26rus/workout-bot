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

// Save возвращает replaced == true, если тренировка этого дня уже была и перезаписана
func (s *DocTabWorkoutService) Save(ctx context.Context, w model.DocTabWorkout) (bool, error) {
	return s.repo.Save(ctx, w)
}

func (s *DocTabWorkoutService) GetAllInfo(ctx context.Context) ([]model.DocTabWorkoutInfo, error) {
	return s.repo.GetAllInfo(ctx)
}

func (s *DocTabWorkoutService) GetFile(ctx context.Context, id int64) ([]byte, error) {
	return s.repo.GetFile(ctx, id)
}

func (s *DocTabWorkoutService) SetTgFileID(ctx context.Context, id int64, tgFileID string) error {
	return s.repo.SetTgFileID(ctx, id, tgFileID)
}
