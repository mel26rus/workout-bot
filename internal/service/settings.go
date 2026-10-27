package service

import (
	"context"
	"fmt"
	"time"

	"workout-bot/internal/model"
	"workout-bot/internal/repository"
)

const keySendTime = "send_time"

type SettingsService struct {
	repo            *repository.SettingsRepo
	defaultSendTime string
}

// NewSettingsService: defaultSendTime (ЧЧ:ММ из конфига) действует, пока время не сохранено в веб-модуле
func NewSettingsService(r *repository.SettingsRepo, defaultSendTime string) *SettingsService {
	return &SettingsService{repo: r, defaultSendTime: defaultSendTime}
}

// ParseSendTime разбирает время рассылки в формате ЧЧ:ММ
func ParseSendTime(value string) (int, int, error) {
	t, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, model.ErrInvalidSendTime
	}

	return t.Hour(), t.Minute(), nil
}

// GetSendTime возвращает время рассылки в формате ЧЧ:ММ
func (s *SettingsService) GetSendTime(ctx context.Context) (string, error) {
	value, ok, err := s.repo.Get(ctx, keySendTime)
	if err != nil {
		return "", err
	}
	if !ok {
		return s.defaultSendTime, nil
	}

	return value, nil
}

func (s *SettingsService) SetSendTime(ctx context.Context, value string) error {
	hour, minute, err := ParseSendTime(value)
	if err != nil {
		return err
	}

	// сохраняем в одном виде, даже если пришло "8:00"
	return s.repo.Set(ctx, keySendTime, fmt.Sprintf("%02d:%02d", hour, minute))
}
