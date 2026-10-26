package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"workout-bot/internal/bot"
	"workout-bot/internal/flow"
	"workout-bot/internal/repository"
	"workout-bot/internal/scheduler"
	"workout-bot/internal/service"
	"workout-bot/internal/web"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// адрес Bot API, переменной сделан для проверки на подставном сервере
var botAPIEndpoint = tgbotapi.APIEndpoint

func (a *App) Boot() error {
	slog.Info("Booting application")

	if a.Config.Web.Password == "" {
		return errors.New("web.password не задан в config.yaml")
	}
	if a.Config.BotKey == "" {
		return errors.New("bot_key не задан в config.yaml")
	}

	loc, err := time.LoadLocation(a.Config.TimeZone)
	if err != nil {
		return fmt.Errorf("time_zone: %w", err)
	}

	sendTime, err := time.Parse("15:04", a.Config.SendTime)
	if err != nil {
		return fmt.Errorf("send_time должно быть в формате ЧЧ:ММ, например \"08:00\": %w", err)
	}

	// bot api
	transport := &http.Transport{}
	if a.Config.BotProxy > "" {
		proxyURL, err := url.Parse(a.Config.BotProxy)
		if err != nil {
			return fmt.Errorf("parse proxy: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	// таймаут больше, чем long polling (30 секунд)
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}

	a.botAPI, err = tgbotapi.NewBotAPIWithClient(a.Config.BotKey, botAPIEndpoint, client)
	if err != nil {
		return fmt.Errorf("bot init failed: %w", err)
	}
	slog.Info("Bot", "Авторизован под аккаунтом", a.botAPI.Self.UserName)
	slog.Info("botApi inited")

	// repos
	userRepo := repository.NewUserRepo(a.DB)
	docWorkoutRepo := repository.NewDocWorkoutRepo(a.DB)
	docTabWorkoutRepo := repository.NewDocTabWorkoutRepo(a.DB)
	docUserWorkoutRepo := repository.NewDocUserWorkoutRepo(a.DB)
	docTabUserWorkoutRepo := repository.NewDocTabUserWorkoutRepo(a.DB)
	slog.Info("Repositories inited")
	// services
	userService := service.NewUserService(userRepo)
	docWorkoutService := service.NewDocWorkoutService(docWorkoutRepo)
	docTabWorkoutService := service.NewDocTabWorkoutService(docTabWorkoutRepo)
	docUserWorkoutService := service.NewDocUserWorkoutService(docUserWorkoutRepo, loc)
	docTabUserWorkoutService := service.NewDocTabUserWorkoutService(docTabUserWorkoutRepo)
	slog.Info("Services inited")
	// flow
	mainFlow := flow.NewMainFlow(
		userService,
		docWorkoutService,
		docTabWorkoutService,
		docTabUserWorkoutService,
	)
	slog.Info("Flows inited")

	a.handler = bot.NewHandler(a.botAPI, mainFlow, loc)

	sch := scheduler.New()

	sch.Add(
		scheduler.NewSendWorkoutJob(a.handler, sendTime.Hour(), sendTime.Minute(), loc),
	)

	a.Scheduler = sch

	a.web, err = web.NewServer(
		a.Config.Web,
		userService,
		docWorkoutService,
		docTabWorkoutService,
		docUserWorkoutService,
	)
	if err != nil {
		return fmt.Errorf("web server: %w", err)
	}
	slog.Info("Web server inited")

	slog.Info("Boot aplication completed")
	return nil
}
