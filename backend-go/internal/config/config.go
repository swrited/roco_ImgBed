package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBType     string
	AppPort    string
	JWTSecret  string
	AppURL     string
}

var appConfig *Config

func Get() *Config {
	if appConfig == nil {
		appConfig = Load()
	}
	return appConfig
}

func Load() *Config {
	_ = godotenv.Load()

	appConfig = &Config{
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USERNAME", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_DATABASE", "lskypro"),
		DBType:     strings.ToLower(getEnv("DB_CONNECTION", "sqlite")),
		AppPort:    getEnv("APP_PORT", "8000"),
		JWTSecret:  getEnv("JWT_SECRET", "lskypro-secret-change-me"),
		AppURL:     getEnv("APP_URL", "http://localhost:8000"),
	}
	return appConfig
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
