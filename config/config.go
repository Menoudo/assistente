package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	MiniMaxAPIKey    string
	DatabaseURL      string
	LogLevel         string
	ServerPort       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	config := &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		MiniMaxAPIKey:    getEnv("MINIMAX_API_KEY", ""),
		DatabaseURL:      getEnv("DATABASE_URL", "./bot.db"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validateConfig(config *Config) error {
	if config.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if config.MiniMaxAPIKey == "" {
		return fmt.Errorf("MINIMAX_API_KEY is required")
	}

	if config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.LogLevel == "debug"
}
