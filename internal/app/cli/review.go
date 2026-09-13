package cli

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
)

type ReviewUseCase interface {
	Exec(ctx context.Context, mergeRequestURL string) error
}

func Review(useCase ReviewUseCase) *cli.Command {
	return &cli.Command{
		Name:   "review",
		Usage:  "review [merge_request_url]",
		Action: ReviewHandler(useCase),
	}
}

func ReviewHandler(useCase ReviewUseCase) cli.ActionFunc {
	return func(ctx context.Context, command *cli.Command) error {
		mergeRequestURL := command.Args().Get(0)
		if mergeRequestURL == "" {
			Help(ctx, command)

			return errors.New("missing required argument: merge request URL")
		}

		return useCase.Exec(ctx, mergeRequestURL)
	}
}
