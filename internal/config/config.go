package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI      string
	MongoDB       string
	AppPort       string
	UploadDir     string
	JWTSecret     string
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		MongoURI:      getEnv("MONGODB_URI", "mongodb://admin:secret@localhost:27017"),
		MongoDB:       getEnv("MONGODB_DATABASE", "syairullah_hajiumroh"),
		AppPort:       getEnv("APP_PORT", "8080"),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
