package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func PyCommand() *cli.Command {
	return &cli.Command{
		Name:  "py",
		Usage: "Manage Python packages",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a Python package",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "dev",
						Usage:   "Add as dev dependency",
						Aliases: []string{"d"},
					},
				},
				Action: func(c context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) == 0 {
						return fmt.Errorf("package name is required")
					}

					packageName := args[0]
					isDev := cmd.Bool("dev")

					return addPythonPackage(packageName, isDev)
				},
			},
			{
				Name:  "remove",
				Usage: "Remove a Python package",
				Action: func(c context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) == 0 {
						return fmt.Errorf("package name is required")
					}

					packageName := args[0]
					return removePythonPackage(packageName)
				},
			},
		},
	}
}

func addPythonPackage(packageName string, isDev bool) error {
	logger := utils.NewLogger()
	executor := utils.NewExecutor()

	projectDir, err := executor.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	if !executor.FileExists(filepath.Join(projectDir, "pyproject.toml")) {
		return fmt.Errorf("pyproject.toml not found - not a Python project")
	}

	if !executor.CommandExists("uv") {
		return fmt.Errorf("uv is not installed. Please install it first")
	}

	logger.Info("Adding Python package: %s (dev: %v)", packageName, isDev)

	ctx := context.Background()
	var args []string

	if isDev {
		args = []string{"add", "--group", "dev", packageName}
	} else {
		args = []string{"add", packageName}
	}

	err = executor.RunCommand(ctx, projectDir, "uv", args...)
	if err != nil {
		return fmt.Errorf("failed to add Python package: %v", err)
	}

	logger.Success("Python package %s added successfully", packageName)
	return nil
}

func removePythonPackage(packageName string) error {
	logger := utils.NewLogger()
	executor := utils.NewExecutor()

	projectDir, err := executor.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	if !executor.FileExists(filepath.Join(projectDir, "pyproject.toml")) {
		return fmt.Errorf("pyproject.toml not found - not a Python project")
	}

	if !executor.CommandExists("uv") {
		return fmt.Errorf("uv is not installed. Please install it first")
	}

	logger.Info("Removing Python package: %s", packageName)

	ctx := context.Background()
	args := []string{"remove", packageName}

	err = executor.RunCommand(ctx, projectDir, "uv", args...)
	if err != nil {
		return fmt.Errorf("failed to remove Python package: %v", err)
	}

	logger.Success("Python package %s removed successfully", packageName)
	return nil
}
