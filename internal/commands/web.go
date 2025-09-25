package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"eel-cli/internal/config"
	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func WebCommand() *cli.Command {
	return &cli.Command{
		Name:  "web",
		Usage: "Manage web packages",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "Add a web package",
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

					return addWebPackage(packageName, isDev)
				},
			},
			{
				Name:  "remove",
				Usage: "Remove a web package",
				Action: func(c context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()
					if len(args) == 0 {
						return fmt.Errorf("package name is required")
					}

					packageName := args[0]
					return removeWebPackage(packageName)
				},
			},
		},
	}
}

func addWebPackage(packageName string, isDev bool) error {
	logger := utils.NewLogger()
	executor := utils.NewExecutor()

	projectDir, err := executor.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	webDir := filepath.Join(projectDir, "web")
	if !executor.DirExists(webDir) {
		return fmt.Errorf("web directory not found")
	}

	manager := cfg.Manager
	if manager == "" {
		manager = detectPackageManager()
	}

	logger.Info("Adding web package: %s (dev: %v)", packageName, isDev)

	ctx := context.Background()
	var args []string

	switch manager {
	case "npm":
		args = []string{"install"}
		if isDev {
			args = append(args, "--save-dev")
		}
		args = append(args, packageName)
		err = executor.RunCommand(ctx, webDir, "npm", args...)

	case "yarn":
		args = []string{"add"}
		if isDev {
			args = append(args, "--dev")
		}
		args = append(args, packageName)
		err = executor.RunCommand(ctx, webDir, "yarn", args...)

	case "pnpm":
		args = []string{"add"}
		if isDev {
			args = append(args, "--save-dev")
		}
		args = append(args, packageName)
		err = executor.RunCommand(ctx, webDir, "pnpm", args...)

	case "bun":
		args = []string{"add"}
		if isDev {
			args = append(args, "--dev")
		}
		args = append(args, packageName)
		err = executor.RunCommand(ctx, webDir, "bun", args...)

	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	if err != nil {
		return fmt.Errorf("failed to add package: %v", err)
	}

	logger.Success("Package %s added successfully", packageName)
	return nil
}

func removeWebPackage(packageName string) error {
	logger := utils.NewLogger()
	executor := utils.NewExecutor()

	projectDir, err := executor.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	cfg, err := config.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	webDir := filepath.Join(projectDir, "web")
	if !executor.DirExists(webDir) {
		return fmt.Errorf("web directory not found")
	}

	manager := cfg.Manager
	if manager == "" {
		manager = detectPackageManager()
	}

	logger.Info("Removing web package: %s", packageName)

	ctx := context.Background()
	var args []string

	switch manager {
	case "npm":
		args = []string{"uninstall", packageName}
		err = executor.RunCommand(ctx, webDir, "npm", args...)

	case "yarn":
		args = []string{"remove", packageName}
		err = executor.RunCommand(ctx, webDir, "yarn", args...)

	case "pnpm":
		args = []string{"remove", packageName}
		err = executor.RunCommand(ctx, webDir, "pnpm", args...)

	case "bun":
		args = []string{"remove", packageName}
		err = executor.RunCommand(ctx, webDir, "bun", args...)

	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	if err != nil {
		return fmt.Errorf("failed to remove package: %v", err)
	}

	logger.Success("Package %s removed successfully", packageName)
	return nil
}
