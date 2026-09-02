package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env  string `env:"ENV" env-required:"true"`
	Port string `env:"PORT" env-required:"true"`

	DBHost      string `env:"DB_HOST" env-required:"true"`
	DB_PORT     string `env:"DB_PORT" env-required:"true"`
	DB_USER     string `env:"DB_USER" env-required:"true"`
	DB_PASSWORD string `env:"DB_PASSWORD" env-required:"true"`
	DB_NAME     string `env:"DB_NAME" env-required:"true"`
	DB_SSLMODE  string `env:"DB_SSLMODE" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to load .env file %v", err)
	}

	return &cfg
}
