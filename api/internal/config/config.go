// Package config centralise le chargement de la configuration de l'API.

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config regroupe tous les paramètres nécessaires au démarrage de l'API
type Config struct {
	Port          string
	DBPath        string
	JWTSecret     string
	JWTExpiration time.Duration
	Env           string
}

// Load lit le fichier .env s'il existe puis construit la configuration à partir des variables d'environnement, avec des valeurs par défaut adaptées

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./data/stock.db"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiration: time.Duration(getEnvInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		Env:           getEnv("ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
