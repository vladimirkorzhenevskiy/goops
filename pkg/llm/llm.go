package llm

import (
	"errors"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   Model
	Timeout time.Duration
}

func (c Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("`baseURL` undefined")
	}

	if c.BaseURL == "" {
		return errors.New("`apiKey` undefined")
	}

	return nil
}

type Client struct {
	cfg Config
	sdk openai.Client
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	sdk := openai.NewClient(
		option.WithBaseURL(cfg.BaseURL),
		option.WithAPIKey(cfg.APIKey),
		option.WithRequestTimeout(cfg.Timeout),
	)

	client := &Client{
		cfg: cfg,
		sdk: sdk,
	}

	return client, nil
}
