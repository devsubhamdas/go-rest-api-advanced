package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv string `env:"APP_ENV" env-required:"true"`
	Port   string `env:"PORT" env-required:"true"`

	// Local or Containerized PostgreSQL config
	DBHost     string `env:"DB_HOST" env-required:"true"`
	DBPort     string `env:"DB_PORT" env-required:"true"`
	DBUser     string `env:"DB_USER" env-required:"true"`
	DBPassword string `env:"DB_PASSWORD" env-required:"true"`
	DBName     string `env:"DB_NAME" env-required:"true"`
	DBSSLMode  string `env:"DB_SSLMODE" env-required:"true"`

	// AllowedOrigins is a comma-separated list in the env, e.g.
	// ALLOWED_ORIGINS=https://yourapp.com,http://localhost:4200
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" env-separator:","`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to read env file %v", err)
	}

	return &cfg
}
