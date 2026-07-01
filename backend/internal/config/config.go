package config

import (
	"fmt"
	"os"
)

// Config хранит все настройки приложения, читаемые из переменных окружения.
type Config struct {
	HTTPAddr    string // адрес, на котором слушает HTTP-сервер, напр. ":8080"
	DatabaseURL string // строка подключения к PostgreSQL
	FrontendDir string // путь к каталогу со статикой фронтенда
}

// Load собирает конфигурацию из переменных окружения, подставляя значения по умолчанию.
func Load() Config {
	cfg := Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		FrontendDir: getEnv("FRONTEND_DIR", "../frontend"),
	}

	if cfg.DatabaseURL == "" {
		// Собираем DSN из отдельных переменных, если целиком не задан.
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "tasks_user")
		pass := getEnv("DB_PASSWORD", "tasks_pass")
		name := getEnv("DB_NAME", "tasks_db")
		sslmode := getEnv("DB_SSLMODE", "disable")

		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, pass, host, port, name, sslmode,
		)
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
