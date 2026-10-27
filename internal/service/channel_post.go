package service

import (
	"context"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

type ChannelPostService struct {
	repo *repository.ChannelPostRepo
}

func NewChannelPostService(r *repository.ChannelPostRepo) *ChannelPostService {
	return &ChannelPostService{repo: r}
}

func (s *ChannelPostService) Save(ctx context.Context, p model.ChannelPost) error {
	return s.repo.Save(ctx, p)
}

func (s *ChannelPostService) GetAllInfo(ctx context.Context) ([]model.ChannelPostInfo, error) {
	return s.repo.GetAllInfo(ctx)
}
