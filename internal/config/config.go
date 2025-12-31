package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpPort string
	HttpUrl  string
}

func GetConfig() *Config {
	if err := godotenv.Load(".env"); err != nil {
		return &Config{
			HttpPort: "localhost",
			HttpUrl:  "8080",
		}
	}

	return &Config{
		HttpPort: os.Getenv("SHORTLINK_HTTP_PORT"),
		HttpUrl:  os.Getenv("SHORTLINK_HTTP_URL"),
	}
}
