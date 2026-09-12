package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/vladimirkorzhenevskiy/goops/pkg/llm"
)

type Config struct {
	LLM    LLM    `envconfig:"LLM"`
	GitLab GitLab `envconfig:"GITLAB"`
}

type LLM struct {
	BaseURL string        `envconfig:"BASE_URL" required:"true"`
	APIKey  string        `envconfig:"API_KEY"  required:"true"`
	Model   llm.Model     `envconfig:"MODEL"    required:"false" default:"deepseek-flash"`
	Timeout time.Duration `envconfig:"TIMEOUT"  required:"false" default:"600s"`
}

type GitLab struct {
	BaseURL string        `envconfig:"BASE_URL" required:"true"`
	APIKey  string        `envconfig:"API_KEY"  required:"true"`
	Timeout time.Duration `envconfig:"TIMEOUT"  required:"false" default:"30s"`
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
