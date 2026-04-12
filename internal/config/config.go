package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	GinMode          string
	MongoURI         string
	MongoDBName      string
	JWTSecret        string
	JWTExpiryHours   int
	OTPExpiryMinutes int
	AppName          string
	AllowedOrigins   string

	// MessageCentral OTP service
	MCCustomerID  string
	MCAuthToken   string
	MCCountryCode string

	// Dev bypass: when true, skip MessageCentral and accept OTP "000000"
	DevOTPBypass bool
}

var AppConfig *Config

func Load() {
	// Try loading .env from several locations so the server works regardless
	// of which directory it is started from.
	_, callerFile, _, _ := runtime.Caller(0)                         // …/internal/config/config.go
	projectRoot := filepath.Join(filepath.Dir(callerFile), "..", "..") // two levels up → project root

	candidates := []string{
		".env",
		filepath.Join(projectRoot, ".env"),
	}

	loaded := false
	for _, p := range candidates {
		if err := godotenv.Load(p); err == nil {
			log.Printf("Loaded config from %s", p)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("No .env file found, reading from environment variables")
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "720"))
	otpExpiry, _ := strconv.Atoi(getEnv("OTP_EXPIRY_MINUTES", "5"))

	AppConfig = &Config{
		Port:             getEnv("PORT", "8080"),
		GinMode:          getEnv("GIN_MODE", "debug"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:      getEnv("MONGO_DB_NAME", "home_care_db"),
		JWTSecret:        getEnv("JWT_SECRET", "home_care_secret_key_change_in_prod"),
		JWTExpiryHours:   jwtExpiry,
		OTPExpiryMinutes: otpExpiry,
		AppName:          getEnv("APP_NAME", "HomeCare"),
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "*"),

		MCCustomerID:  getEnv("MC_CUSTOMER_ID", "C-0EAA579E91B1418"),
		MCAuthToken:   getEnv("MC_AUTH_TOKEN", "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJDLTBFQUE1NzlFOTFCMTQxOCIsImlhdCI6MTc3NTk2MzA2NywiZXhwIjoxOTMzNjQzMDY3fQ.J9_N6D8tjoFalF9DqAEo4tcQvAyOpSBFV3hH36ZLOoKeLIbXZPr9O284O5Tv0oiK4tg61WRZB0d4kAIFmwdupA"),
		MCCountryCode: getEnv("MC_COUNTRY_CODE", "91"),
		DevOTPBypass:  getEnv("DEV_OTP_BYPASS", "true") == "true",
	}

	log.Printf("[Config] MC_CUSTOMER_ID=%s | MC_AUTH_TOKEN=%s...%s",
		AppConfig.MCCustomerID,
		AppConfig.MCAuthToken[:10],
		AppConfig.MCAuthToken[len(AppConfig.MCAuthToken)-6:],
	)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
