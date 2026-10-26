package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "time/tzdata"

	"workout-bot/internal/app"
	"workout-bot/internal/config"
	"workout-bot/internal/db"
	"workout-bot/internal/logger"
)

var configPath = flag.String(
	"config",
	"config.yaml",
	"Path to config file",
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	err = logger.New(cfg)
	if err != nil {
		panic(err)
	}
	slog.Info("Logger inited")

	dbPool := db.NewPostgres(cfg.Database.URL)
	slog.Info("dbPool inited")
	defer dbPool.Close()
	defer slog.Info("dbPool closed")

	application := &app.App{
		Config: cfg,
		DB:     dbPool,
		Logger: slog.Default(),
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	application.Run(ctx)

	slog.Info("App closed")
}
