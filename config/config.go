package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	AuthServiceURL string
	AuthIssuer     string
	WhatsApp       WhatsAppConfig
	Email          EmailConfig
}

type EmailConfig struct {
	ResendAPIKey string
	FromEmail    string
	FrontendURL  string
}

type WhatsAppConfig struct {
	Mode         string // "mock" or "evolution"
	APIURL       string
	APIKey       string
	InstanceName string
}

func Load() (*Config, error) {
	// Try to load .env file (ignore error if not exists)
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", ""),
		AuthIssuer:     getEnv("AUTH_ISSUER", ""),
		WhatsApp: WhatsAppConfig{
			Mode:         getEnv("WHATSAPP_MODE", "mock"),
			APIURL:       getEnv("EVOLUTION_API_URL", ""),
			APIKey:       getEnv("EVOLUTION_API_KEY", ""),
			InstanceName: getEnv("EVOLUTION_INSTANCE_NAME", ""),
		},
		Email: EmailConfig{
			ResendAPIKey: getEnv("RESEND_API_KEY", ""),
			FromEmail:    getEnv("FROM_EMAIL", ""),
			FrontendURL:  getEnv("FRONTEND_URL", ""),
		},
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	origin, err := url.Parse(cfg.Email.FrontendURL)
	if err != nil || origin.Host == "" || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || (origin.Path != "" && origin.Path != "/") || origin.RawQuery != "" || origin.Fragment != "" {
		return nil, fmt.Errorf("FRONTEND_URL must be the frontend origin (scheme and host, no path)")
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
