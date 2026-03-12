package config

import (
	"os"
)

type Config struct {
	ServerPort         string
	AlertmanagerURL    string
	AlertmanagerToggle bool
	DBPath             string
}

func Load() *Config {
	return &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		AlertmanagerURL:    getEnv("ALERTMANAGER_URL", ""),
		AlertmanagerToggle: getEnv("ALERTMANAGER_URL", "") != "",
		DBPath:             getEnv("DB_PATH", "/data/servicepatrol.db"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}
