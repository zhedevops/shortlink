package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPURL  string
	HTTPPort string
}

func GetConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		return &Config{
			HTTPURL:  "localhost",
			HTTPPort: "8080",
		}
	}

	return &Config{
		HTTPURL:  os.Getenv("SHORTLINK_HTTP_URL"),
		HTTPPort: os.Getenv("SHORTLINK_HTTP_PORT"),
	}
}
