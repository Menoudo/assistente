package config

import (
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	TelegramBotToken string
	MiniMaxAPIKey    string
	DatabaseURL      string
	LogLevel         string
	ServerPort       string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config := &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		MiniMaxAPIKey:    getEnv("MINIMAX_API_KEY", ""),
		DatabaseURL:      getEnv("DATABASE_URL", "./bot.db"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("конфигурация невалидна: %w", err)
	}

	return config, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateConfig проверяет обязательные переменные конфигурации
func validateConfig(config *Config) error {
	if config.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN обязателен")
	}

	if config.MiniMaxAPIKey == "" {
		return fmt.Errorf("MINIMAX_API_KEY обязателен")
	}

	if config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL обязателен")
	}

	return nil
}

// IsDevelopment проверяет, запущено ли приложение в режиме разработки
func (c *Config) IsDevelopment() bool {
	return c.LogLevel == "debug"
}
