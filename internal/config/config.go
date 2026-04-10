package config

import (
	"log"
	"os"
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
	OTPTestMode      bool
	OTPTestValue     string
	AppName          string
	AllowedOrigins   string
}

var AppConfig *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "720"))
	otpExpiry, _ := strconv.Atoi(getEnv("OTP_EXPIRY_MINUTES", "5"))
	otpTestMode, _ := strconv.ParseBool(getEnv("OTP_TEST_MODE", "false"))

	AppConfig = &Config{
		Port:             getEnv("PORT", "8080"),
		GinMode:          getEnv("GIN_MODE", "debug"),
		MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:      getEnv("MONGO_DB_NAME", "home_care_db"),
		JWTSecret:        getEnv("JWT_SECRET", "home_care_secret_key"),
		JWTExpiryHours:   jwtExpiry,
		OTPExpiryMinutes: otpExpiry,
		OTPTestMode:      otpTestMode,
		OTPTestValue:     getEnv("OTP_TEST_VALUE", "5555"),
		AppName:          getEnv("APP_NAME", "HomeCare"),
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
