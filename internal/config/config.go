package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv" // Optional: For loading .env files during local development
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	ServerPort int
	DynamoDB   DynamoDBConfig
	Auth       AuthConfig
	LogLevel   string
	// Add other application-specific configurations here
}

// DynamoDBConfig holds DynamoDB specific configurations.
type DynamoDBConfig struct {
	AWSRegion        string
	TableName        string
	EntityTypeIndex  string
	UseDynamoDBLocal bool   // To switch to DynamoDB Local for testing/development
	DynamoDBLocalURL string // URL for DynamoDB Local (e.g., http://localhost:8000)
	// You might add Read/Write capacity settings if using provisioned mode and managing it here
}

// AuthConfig holds authentication-related configurations (e.g., JWT secrets).
type AuthConfig struct {
	JWTSecret      string
	TokenExpiryHrs int
}

// LoadConfig loads application configuration from environment variables.
// It's good practice to call this once at application startup.
func LoadConfig() (*AppConfig, error) {
	// Optional: Load .env file for local development.
	// In production, environment variables are usually set directly.
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load() // Loads .env from the current directory
		if err != nil {
			log.Println("Warning: .env file not found or error loading it:", err)
		}
	}

	appCfg := &AppConfig{
		ServerPort: getEnvAsInt("SERVER_PORT", 8080),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		DynamoDB: DynamoDBConfig{
			AWSRegion:        getEnv("AWS_REGION", "us-east-1"), // Default to a common region
			TableName:        getEnv("DYNAMODB_TABLE_NAME", "Resources"),
			EntityTypeIndex:  getEnv("DYNAMODB_ENTITY_TYPE_GSI_NAME", "EntityTypeIndex"),
			UseDynamoDBLocal: getEnvAsBool("DYNAMODB_USE_LOCAL", false),
			DynamoDBLocalURL: getEnv("DYNAMODB_LOCAL_URL", "http://localhost:8000"),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", "a_very_secure_secret_key_please_change_me"), // CHANGE THIS!
			TokenExpiryHrs: getEnvAsInt("JWT_TOKEN_EXPIRY_HOURS", 24),
		},
	}

	// Validate essential configurations if necessary
	if appCfg.Auth.JWTSecret == "a_very_secure_secret_key_please_change_me" && os.Getenv("APP_ENV") == "production" {
		log.Println("CRITICAL WARNING: Default JWT_SECRET is being used in a production-like environment. Please set a strong, unique secret.")
	}
	if appCfg.DynamoDB.TableName == "" {
		log.Fatal("FATAL: DYNAMODB_TABLE_NAME environment variable is not set.")
	}

	return appCfg, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool retrieves an environment variable as a boolean or returns a default.
// Considers "true", "1", "yes" as true (case-insensitive).
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(valueStr)
	if err == nil {
		return val
	}
	// Fallback for "yes" or "1" if ParseBool fails (it's strict)
	lcVal := strings.ToLower(valueStr)
	if lcVal == "yes" || lcVal == "1" || lcVal == "true" {
		return true
	}
	return defaultValue
}
