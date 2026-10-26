package service

import (
	"context"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type UserService struct {
	repo *repository.UserRepo
}

func NewUserService(r *repository.UserRepo) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) GetOrCreate(ctx context.Context, u model.User) (bool, error) {
	return s.repo.GetOrCreate(ctx, u)
}

func (s *UserService) SetDocWorkout(ctx context.Context, tgID int64, docWorkoutID int64) error {
	return s.repo.SetDocWorkout(ctx, tgID, docWorkoutID)
}

func (s *UserService) GetAdminTgIDs(ctx context.Context) ([]int64, error) {
	return s.repo.GetAdminTgIDs(ctx)
}

func (s *UserService) GetParticipants(ctx context.Context) ([]model.Participant, error) {
	return s.repo.GetParticipants(ctx)
}
