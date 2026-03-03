package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI   string
	MongoDB    string
	AppPort    string
	UploadDir  string
}

func Load() *Config {
	godotenv.Load()

	cfg := &Config{
		MongoURI: getEnv("MONGODB_URI", "mongodb://admin:secret@localhost:27017"),
		MongoDB:  getEnv("MONGODB_DATABASE", "syairullah_hajiumroh"),
		AppPort:   getEnv("APP_PORT", "8080"),
		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
