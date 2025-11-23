package config

import (
	"fmt"
	"os"
)

// Config хранит конфигурацию приложения.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  string
}

// NewFromEnv создает новую конфигурацию из переменных окружения.
func NewFromEnv() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "reviewer_user"),
		DBPassword: getEnv("DB_PASSWORD", "reviewer_password"),
		DBName:     getEnv("DB_NAME", "reviewer_db"),
		DBSslMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// DSN возвращает строку подключения для pgx (с пробелами).
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSslMode)
}

// MigrationURL возвращает строку подключения в формате URL для библиотеки migrate.
func (c *Config) MigrationURL() string {
	// postgresql://user:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSslMode)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
