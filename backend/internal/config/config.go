package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Address string `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
}

type Config struct {
	ENV    string `env:"ENV" env-default:"dev"`
	DBPath string `env:"DB_PATH" env-default:"sqlite/dev"`
	DBName string `env:"DB_NAME" env-default:"api.db"`
	HTTPServer
	JWTKey string `env:"JWT_KEY" env-default:"supersecretjwtkey"`
}

func LoadConfig() *Config {
	var cfg Config
	var envPath string

	flag.StringVar(&envPath, "config", "", "path to .env file")
	flag.Parse()

	if envPath == "" {
		envPath = os.Getenv("CONFIG_PATH")
	}

	if envPath == "" {
		envPath = "config/dev.env"
	}

	if err := cleanenv.ReadConfig(envPath, &cfg); err != nil {
		log.Fatalf("Failed to read config from %s, err: %v", envPath, err)
	}

	return &cfg
}
