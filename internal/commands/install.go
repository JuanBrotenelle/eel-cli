package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"eel-cli/internal/config"
	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func InstallCommand() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install project dependencies (web packages, uv, and create eel.d.ts)",
		Action: func(c context.Context, cmd *cli.Command) error {
			return installDependencies()
		},
	}
}

func installDependencies() error {
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

	if !executor.FileExists(filepath.Join(projectDir, "main.py")) {
		return fmt.Errorf("not in an Eel project directory (main.py not found)")
	}

	logger.Info("Installing dependencies...")

	if !executor.CommandExists("uv") {
		logger.Info("Installing uv...")
		if err := installUV(); err != nil {
			logger.Warning("Failed to install uv: %v", err)
		} else {
			logger.Success("uv installed successfully")
		}
	} else {
		logger.Info("uv is already installed")
	}

	logger.Info("Installing Python dependencies...")
	ctx := context.Background()
	if err := executor.RunCommand(ctx, projectDir, "uv", "sync"); err != nil {
		return fmt.Errorf("failed to install Python dependencies: %v", err)
	}
	logger.Success("Python dependencies installed")

	webDir := filepath.Join(projectDir, "web")
	if executor.DirExists(webDir) {
		logger.Info("Installing web dependencies...")

		manager := cfg.Manager
		if manager == "" {
			manager = detectPackageManager()
		}

		if err := installWebDependencies(webDir, manager); err != nil {
			return fmt.Errorf("failed to install web dependencies: %v", err)
		}
		logger.Success("Web dependencies installed")
	}

	if err := createEelTypes(); err != nil {
		logger.Warning("Failed to create eel.d.ts: %v", err)
	} else {
		logger.Success("Created eel.d.ts")
	}

	logger.Success("All dependencies installed successfully!")
	return nil
}

func installUV() error {
	executor := utils.NewExecutor()
	ctx := context.Background()

	if err := executor.RunCommandSilent(ctx, "", "powershell", "-c", "irm https://astral.sh/uv/install.ps1 | iex"); err != nil {
		if err := executor.RunCommandSilent(ctx, "", "curl", "-LsSf", "https://astral.sh/uv/install.sh", "|", "sh"); err != nil {
			return fmt.Errorf("failed to install uv: %v", err)
		}
	}

	return nil
}

func detectPackageManager() string {
	executor := utils.NewExecutor()

	managers := []string{"bun", "pnpm", "yarn", "npm"}
	for _, manager := range managers {
		if executor.CommandExists(manager) {
			return manager
		}
	}

	return "npm"
}

func installWebDependencies(webDir, manager string) error {
	executor := utils.NewExecutor()
	ctx := context.Background()

	switch manager {
	case "npm":
		return executor.RunCommand(ctx, webDir, "npm", "install")
	case "yarn":
		return executor.RunCommand(ctx, webDir, "yarn", "install")
	case "pnpm":
		return executor.RunCommand(ctx, webDir, "pnpm", "install")
	case "bun":
		return executor.RunCommand(ctx, webDir, "bun", "install")
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func createEelTypes() error {
	executor := utils.NewExecutor()
	projectDir, err := executor.GetWorkingDir()
	if err != nil {
		return err
	}

	webDir := filepath.Join(projectDir, "web")
	if !executor.DirExists(webDir) {
		return fmt.Errorf("web directory not found")
	}

	eelTypes := `// Eel type definitions
declare namespace eel {
  function expose(func: Function, name?: string): void;
  function start(path: string, options?: {
    size?: [number, number];
    port?: number;
    host?: string;
    mode?: string;
    block?: boolean;
    close_callback?: Function;
    shutdown_delay?: number;
    [key: string]: any;
  }): void;
  function init(path: string): void;
}

// Global eel object
declare const eel: typeof eel;
`

	typesPath := filepath.Join(webDir, "eel.d.ts")
	return os.WriteFile(typesPath, []byte(eelTypes), 0644)
}
