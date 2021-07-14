package app

import (
	"os"
)

var (
	SecretKey   = GetEnv("SECRET_KEY", "")
	DatabaseURL = GetEnv("DATABASE_URL", "postgres://localhost/kakadoo?sslmode=disable")
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
