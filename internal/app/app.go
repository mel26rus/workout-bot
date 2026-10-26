package app

import (
	"context"
	"log/slog"
	"sync"

	"workout-bot/internal/bot"
	"workout-bot/internal/config"
	"workout-bot/internal/scheduler"
	"workout-bot/internal/web"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config    *config.AppConfig
	Logger    *slog.Logger
	DB        *pgxpool.Pool
	Scheduler *scheduler.Scheduler
	botAPI    *tgbotapi.BotAPI
	handler   *bot.Handler
	web       *web.Server
}

// тут всё запускаем
func (a *App) Run(ctx context.Context) {
	a.Logger.Info("+App.run")
	a.Logger.Debug("+App.boot")
	if err := a.Boot(); err != nil {
		a.Logger.Error("App.boot", "Error", err)
		return
	}
	a.Logger.Debug("-App.boot")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.web.Run(ctx)
	}()
	a.Logger.Info("Web started")

	go a.Scheduler.Run(ctx)
	a.Logger.Info("Scheduler started")

	a.Logger.Info("+handler.Start(ctx)")
	a.handler.Start(ctx)
	a.Logger.Info("-handler.Start(ctx)")

	wg.Wait()
	a.Logger.Info("-App.run")
}
