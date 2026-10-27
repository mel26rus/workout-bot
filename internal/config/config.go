package config

import (
	"fmt"
	"os"
	"path/filepath"

	"workout-bot/internal/filesystem"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	URL string `yaml:"url"`
}
type LogConfig struct {
	Folder  string `yaml:"folder"`
	Level   string `yaml:"level"`
	Console bool   `yaml:"console"`
	File    bool   `yaml:"file"`
}
type WebConfig struct {
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type AppConfig struct {
	LogConfig LogConfig      `yaml:"log_config"`
	Database  DatabaseConfig `yaml:"database"`
	Web       WebConfig      `yaml:"web"`
	BotKey    string         `yaml:"bot_key"`
	BotProxy  string         `yaml:"bot_proxy"`
	ChannelID int64          `yaml:"channel_id"` // закрытый канал тренера, откуда берутся тренировки
	SendTime  string         `yaml:"send_time"`  // время рассылки по умолчанию, ЧЧ:ММ; тренер меняет его в веб-модуле
	TimeZone  string         `yaml:"time_zone"`
	AppName   string         `yaml:"app_name"`

	// папка, где лежит config.yaml (в самом файле не хранится)
	BaseDir string `yaml:"-"`
}

const DefaultDatabaseURL = "postgres://postgres:root@localhost:5432/workout_db"

func getConfigFile(fileName string) string {
	return filepath.Join(
		filesystem.BaseDir(fileName),
		fileName,
	)
}

func defaultConfig() *AppConfig {
	return &AppConfig{
		LogConfig: LogConfig{
			Folder:  "Logs",
			Level:   "debug",
			Console: true,
			File:    true,
		},
		Database: DatabaseConfig{
			URL: DefaultDatabaseURL,
		},
		Web: WebConfig{
			Port:     8080,
			Password: "",
		},
		BotKey:    "",
		BotProxy:  "",
		ChannelID: 0,
		SendTime:  "08:00",
		TimeZone:  "Europe/Moscow",
		AppName:   "WorkoutBot",
	}
}

func Load(fileName string) (*AppConfig, error) {

	configFile := getConfigFile(fileName)
	data, err := os.ReadFile(configFile)
	if err != nil {
		defConfig := defaultConfig()
		defConfig.BaseDir = filepath.Dir(configFile)
		if saveErr := Save(defConfig, configFile); saveErr != nil {
			return defConfig, saveErr
		}
		return defConfig, fmt.Errorf("конфиг не найден, создан %s с настройками по умолчанию: заполните его (web.password, bot_key, channel_id) и запустите снова", configFile)
	}

	var cfg AppConfig

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.BaseDir = filepath.Dir(configFile)

	return &cfg, nil
}

func Save(cfg *AppConfig, configFile string) error {

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}
