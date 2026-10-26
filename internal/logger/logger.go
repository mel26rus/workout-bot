package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"workout-bot/internal/config"
)

func New(cfg *config.AppConfig) error {

	writers := make([]io.Writer, 0, 2)

	if cfg.LogConfig.Console {
		writers = append(writers, os.Stdout)
	}

	if cfg.LogConfig.File {

		logDir := filepath.Join(
			cfg.BaseDir,
			cfg.LogConfig.Folder,
		)

		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("create log directory: %w", err)
		}

		fileName := fmt.Sprintf(
			"%s_%s.log",
			cfg.AppName,
			time.Now().Format("20060102_150405"),
		)

		filePath := filepath.Join(
			logDir,
			fileName,
		)

		file, err := os.OpenFile(
			filePath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)

		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}

		writers = append(writers, file)
	}

	level := parseLevel(cfg.LogConfig.Level)

	handler := slog.NewTextHandler(
		io.MultiWriter(writers...),
		&slog.HandlerOptions{
			Level: level,
		},
	)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return nil
}

func parseLevel(level string) slog.Level {

	switch level {

	case "debug":
		return slog.LevelDebug

	case "warn":
		return slog.LevelWarn

	case "error":
		return slog.LevelError

	default:
		return slog.LevelInfo
	}
}
