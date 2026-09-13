package app

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/vladimirkorzhenevskiy/goops/internal/app/cli"
	"github.com/vladimirkorzhenevskiy/goops/internal/app/config"
)

func Run(args ...string) error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	deps := &Dependencies{}

	if err = deps.Resolve(ctx, cfg); err != nil {
		return err
	}

	cmd := cli.Root(
		cli.Review(deps.UseCases.Review),
	)

	return cmd.Run(ctx, args)
}
