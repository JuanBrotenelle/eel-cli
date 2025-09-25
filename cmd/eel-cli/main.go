package main

import (
	"context"
	"fmt"
	"os"

	"eel-cli/internal/commands"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "🐍 eel-cli",
		Usage: "CLI utility for creating and managing Eel projects",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "help",
				Usage:   "show help",
				Aliases: []string{"h"},
			},
		},
		Commands: []*cli.Command{
			commands.CreateCommand(),
			commands.InstallCommand(),
			commands.WebCommand(),
			commands.PyCommand(),
			commands.DevCommand(),
			commands.BuildCommand(),
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			fmt.Println("🐍 eel-cli: Use --help or -h to see available commands")
			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "🐍 eel-cli: Error: %v\n", err)
		os.Exit(1)
	}
}
