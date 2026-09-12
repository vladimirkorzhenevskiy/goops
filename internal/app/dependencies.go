package app

import (
	"context"

	"github.com/vladimirkorzhenevskiy/goops/internal/agents/reviewer"
	"github.com/vladimirkorzhenevskiy/goops/internal/app/config"
	"github.com/vladimirkorzhenevskiy/goops/pkg/gitlab"
	"github.com/vladimirkorzhenevskiy/goops/pkg/llm"
)

type Dependencies struct {
	Adapters struct {
		LLM    *llm.Client
		GitLab *gitlab.Client
	}
	UseCases struct {
		Review *reviewer.ReviewUseCase
	}
}

func (d *Dependencies) Resolve(_ context.Context, cfg *config.Config) error {
	{
		adapter, err := llm.New(llm.Config{
			BaseURL: cfg.LLM.BaseURL,
			APIKey:  cfg.LLM.APIKey,
			Model:   cfg.LLM.Model,
			Timeout: cfg.LLM.Timeout,
		})
		if err != nil {
			return err
		}

		d.Adapters.LLM = adapter
	}

	{
		adapter, err := gitlab.New(gitlab.Config{
			BaseURL: cfg.GitLab.BaseURL,
			APIKey:  cfg.GitLab.APIKey,
			Timeout: cfg.GitLab.Timeout,
		})
		if err != nil {
			return err
		}

		d.Adapters.GitLab = adapter
	}

	{
		useCase, err := reviewer.NewReviewUseCase(
			d.Adapters.GitLab,
			d.Adapters.LLM,
		)
		if err != nil {
			return err
		}

		d.UseCases.Review = useCase
	}

	return nil
}
