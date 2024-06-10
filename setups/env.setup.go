package setups

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/marine-br/golib-logger/logger"
	"os"
)

var errors []string

func SetupEnv() {
	defaultEnv("LOGGER_SERVICE_NAME", "geofence-service")

	validateEnv("MONGO_DB")
	validateEnv("MONGO_URI")

	validateEnv("RABBITMQ_NAME")
	validateEnv("RABBITMQ_HOST")
	validateEnv("RABBITMQ_PASSWORD")
	validateEnv("RABBITMQ_USERNAME")
	validateEnv("RABBITMQ_NAME")

	defaultEnv("HTTP_SERVER_PORT", ":8080")

	for _, err := range errors {
		logger.Log(err)
	}

	if len(errors) > 0 {
		os.Exit(0)
	}
}

func validateEnv(envName string) {
	env := os.Getenv(envName)
	if env == "" {
		errors = append(errors, "no env", envName)
	}
}

func defaultEnv(envName, defaultValue string) {
	env := os.Getenv(envName)
	if env == "" {
		os.Setenv(envName, defaultValue)
	}
}
