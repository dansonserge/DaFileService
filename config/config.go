package config

import (
	"os"
)

type Config struct {
	Port           string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
	DefaultBucket  string
	AuthServiceURL string
}

func LoadConfig() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "admin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "password"),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		DefaultBucket:  getEnv("DEFAULT_BUCKET", "default-storage"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
