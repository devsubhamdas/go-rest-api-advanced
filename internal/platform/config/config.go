package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string `env:"ENV" env-required:"true"`
	Port string `env:"PORT" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load env file: %v", err)
		return nil
	}

	env := os.Getenv("ENV")
	port := os.Getenv("PORT")

	return &Config{
		Env:  env,
		Port: port,
	}
}
