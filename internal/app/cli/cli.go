package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func Root(commands ...*cli.Command) *cli.Command {
	return &cli.Command{
		Name:     "github.com/vladimirkorzhenevskiy/goops",
		Usage:    "github.com/vladimirkorzhenevskiy/goops [command]",
		Commands: commands,
	}
}

func Help(ctx context.Context, command *cli.Command) {
	_ = cli.ShowCommandHelp(ctx, command, "")
}
