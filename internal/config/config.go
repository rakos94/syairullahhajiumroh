package config

import (
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Defaults that are published in .env.example and must not be used in production.
const (
	DefaultJWTSecret     = "change-me-in-production"
	DefaultAdminPassword = "admin123"
	defaultMongoPassword = "secret"
	minJWTSecretLength   = 32
)

type Config struct {
	MongoURI      string
	MongoDB       string
	AppPort       string
	UploadDir     string
	JWTSecret     string
	AdminUsername string
	AdminPassword string

	// Production is true when GIN_MODE=release (set in the Docker image).
	Production bool
	// AdminResetPassword forces the super admin password back to
	// ADMIN_PASSWORD on startup (account recovery). Normally the password
	// from .env is only used when the super admin is first created.
	AdminResetPassword bool
	// TrustedProxies are IPs/CIDRs allowed to set X-Forwarded-For
	// (e.g. the Caddy container). Empty = trust no proxy.
	TrustedProxies []string
	// EnableSwagger serves /swagger. Off by default in production.
	EnableSwagger bool
}

func Load() *Config {
	godotenv.Load()

	production := os.Getenv("GIN_MODE") == "release"
	cfg := &Config{
		MongoURI:           getEnv("MONGODB_URI", "mongodb://admin:secret@localhost:27017"),
		MongoDB:            getEnv("MONGODB_DATABASE", "syairullah_hajiumroh"),
		AppPort:            getEnv("APP_PORT", "8080"),
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		JWTSecret:          getEnv("JWT_SECRET", DefaultJWTSecret),
		AdminUsername:      getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:      getEnv("ADMIN_PASSWORD", DefaultAdminPassword),
		Production:         production,
		AdminResetPassword: getBool("ADMIN_RESET_PASSWORD", false),
		TrustedProxies:     getList("TRUSTED_PROXIES"),
		EnableSwagger:      getBool("ENABLE_SWAGGER", !production),
	}
	return cfg
}

// SecretProblems lists insecure settings. Fatal ones must stop the app in
// production; the rest are warnings.
func (c *Config) SecretProblems() (fatal, warnings []string) {
	if c.JWTSecret == DefaultJWTSecret {
		fatal = append(fatal, "JWT_SECRET masih default; siapa pun bisa membuat token admin palsu. Buat dengan: openssl rand -hex 32")
	} else if len(c.JWTSecret) < minJWTSecretLength {
		fatal = append(fatal, "JWT_SECRET terlalu pendek (minimal 32 karakter). Buat dengan: openssl rand -hex 32")
	}
	if u, err := url.Parse(c.MongoURI); err == nil && u.User != nil {
		if pw, _ := u.User.Password(); pw == defaultMongoPassword {
			warnings = append(warnings, "Password MongoDB masih default (\"secret\"). Lihat README bagian keamanan untuk menggantinya.")
		}
	}
	return fatal, warnings
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	}
	return fallback
}

func getList(key string) []string {
	var out []string
	for _, p := range strings.Split(os.Getenv(key), ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
