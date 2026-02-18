package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        string
	Environment string
	Debug       bool

	// Database
	DBPath               string
	DBMaxIdleConnections int
	DBMaxOpenConnections int
	DBType               string // sqlite or postgres
	DBHost               string
	DBPort               int
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string

	// Twilio
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioWhatsAppNumber   string
	TwilioAuthTokenConfirm string

	// JWT
	JWTSecret    string
	JWTExpiryHrs int

	// M-Pesa (Future)
	MPesaConsumerKey    string
	MPesaConsumerSecret string
	MPesaShortcode      string
	MPesaPasskey        string
	MPesaEnvironment    string
	MPesaCallbackURL    string

	// OpenAI
	OpenAIAPIKey string

	// Redis
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Rate Limiting
	RateLimitEnabled       bool
	RateLimitMaxRequests   int
	RateLimitWindowSeconds int

	// Logging
	LogLevel string
	LogFile  string

	// Security
	AllowedOrigins string
	CORSEnabled    bool

	// External Services
	AfricaTalkingAPIKey    string
	AfricaTalkingUsername  string
	AfricaTalkingShortCode string
	SendGridAPIKey         string
	SendGridFromEmail      string
	SendGridFromName       string

	// Feature Flags
	FeatureMpesaEnabled         bool
	FeatureAnalyticsEnabled     bool
	FeatureWebDashboardEnabled  bool
	FeatureMultipleShopsEnabled bool
	FeatureStaffAccountsEnabled bool
}

// Load loads configuration from environment variables
// It also loads from .env file if it exists
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := &Config{
		// Server
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Debug:       getEnvAsBool("DEBUG", false),

		// Database
		DBPath:               getEnv("DB_PATH", "./dukapos.db"),
		DBMaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
		DBMaxOpenConnections: getEnvAsInt("DB_MAX_OPEN_CONNECTIONS", 100),
		DBType:               getEnv("DB_TYPE", "sqlite"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnvAsInt("DB_PORT", 5432),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBName:               getEnv("DB_NAME", "dukapos"),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),

		// Twilio
		TwilioAccountSID:       getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:        getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioWhatsAppNumber:   getEnv("TWILIO_WHATSAPP_NUMBER", "whatsapp:+14155238886"),
		TwilioAuthTokenConfirm: getEnv("TWILIO_AUTHENTICATION_TOKEN", ""),

		// JWT
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHrs: getEnvAsInt("JWT_EXPIRY_HOURS", 72),

		// M-Pesa
		MPesaConsumerKey:    getEnv("MPESA_CONSUMER_KEY", ""),
		MPesaConsumerSecret: getEnv("MPESA_CONSUMER_SECRET", ""),
		MPesaShortcode:      getEnv("MPESA_SHORTCODE", ""),
		MPesaPasskey:        getEnv("MPESA_PASSKEY", ""),
		MPesaEnvironment:    getEnv("MPESA_ENVIRONMENT", "sandbox"),
		MPesaCallbackURL:    getEnv("MPESA_CALLBACK_URL", ""),

		// OpenAI
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		// Redis
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// Rate Limiting
		RateLimitEnabled:       getEnvAsBool("RATE_LIMIT_ENABLED", true),
		RateLimitMaxRequests:   getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 100),
		RateLimitWindowSeconds: getEnvAsInt("RATE_LIMIT_WINDOW_SECONDS", 60),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		LogFile:  getEnv("LOG_FILE", "./logs/dukapos.log"),

		// Security
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		CORSEnabled:    getEnvAsBool("CORS_ENABLED", true),

		// External Services
		AfricaTalkingAPIKey:    getEnv("AFRICA_TALKING_API_KEY", ""),
		AfricaTalkingUsername:  getEnv("AFRICA_TALKING_USERNAME", "sandbox"),
		AfricaTalkingShortCode: getEnv("AFRICA_TALKING_SHORT_CODE", ""),
		SendGridAPIKey:         getEnv("SENDGRID_API_KEY", ""),
		SendGridFromEmail:      getEnv("SENDGRID_FROM_EMAIL", "noreply@dukapos.com"),
		SendGridFromName:       getEnv("SENDGRID_FROM_NAME", "DukaPOS"),

		// Feature Flags
		FeatureMpesaEnabled:         getEnvAsBool("FEATURE_MPESA_ENABLED", false),
		FeatureAnalyticsEnabled:     getEnvAsBool("FEATURE_ANALYTICS_ENABLED", true),
		FeatureWebDashboardEnabled:  getEnvAsBool("FEATURE_WEB_DASHBOARD_ENABLED", true),
		FeatureMultipleShopsEnabled: getEnvAsBool("FEATURE_MULTIPLE_SHOPS_ENABLED", false),
		FeatureStaffAccountsEnabled: getEnvAsBool("FEATURE_STAFF_ACCOUNTS_ENABLED", false),
	}

	// Validate required fields
	if cfg.TwilioAccountSID == "" {
		fmt.Println("Warning: TWILIO_ACCOUNT_SID not set")
	}
	if cfg.TwilioAuthToken == "" {
		fmt.Println("Warning: TWILIO_AUTH_TOKEN not set")
	}
	if cfg.JWTSecret == "change-me-in-production" {
		fmt.Println("Warning: Using default JWT_SECRET - change in production!")
	}

	return cfg, nil
}

// GetAllowedOrigins returns the allowed origins as a slice
func (c *Config) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "*" {
		return []string{"*"}
	}
	return strings.Split(c.AllowedOrigins, ",")
}

// GetJWTDuration returns the JWT expiry duration
func (c *Config) GetJWTDuration() time.Duration {
	return time.Duration(c.JWTExpiryHrs) * time.Hour
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetTwilioWebhookURL returns the webhook URL for Twilio
func (c *Config) GetTwilioWebhookURL(baseURL string) string {
	return fmt.Sprintf("%s/webhook/twilio", baseURL)
}

// getEnv gets an environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as int
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
