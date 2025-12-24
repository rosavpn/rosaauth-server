package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port            string
	DBUrl           string
	JWTSecret       string
	LogLevel        string
	AdminEmail      string
	AdminPassword   string
	GlobalRateLimit int
	LoginRateLimit  int
}

func LoadConfig() *Config {
	return &Config{
		Port:            getEnv("PORT", "3000"),
		DBUrl:           getEnvOrPanic("DB_URL"),
		JWTSecret:       getEnvOrPanic("JWT_SECRET"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		AdminEmail:      getEnvOrPanic("ADMIN_EMAIL"),
		AdminPassword:   getEnvOrPanic("ADMIN_PASSWORD"),
		GlobalRateLimit: getEnvAsInt("GLOBAL_RATE_LIMIT", 20),
		LoginRateLimit:  getEnvAsInt("LOGIN_RATE_LIMIT", 5),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvOrPanic(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic("Required environment variable not set: " + key)
	}
	return value
}

func (c *Config) IsDebug() bool {
	return strings.ToLower(c.LogLevel) == "debug"
}
