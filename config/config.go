package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DBHost       string `env:"DB_HOST,required"`
	DBPort       string `env:"DB_PORT,required"`
	DBUser       string `env:"DB_USER,required"`
	DBPass       string `env:"DB_PASS,required"`
	DBName       string `env:"DB_NAME,required"`
	OllamaURL    string `env:"OLLAMA_URL,required"`
	SourceDir    string `env:"SOURCE_DIR,required"`
	ProcessedDir string `env:"PROCESSED_DIR,required"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	return cfg, nil
}
