package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eel-cli/internal/config"
	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func BuildCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build the application using PyInstaller",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Usage:   "Application name",
				Aliases: []string{"n"},
			},
			&cli.StringFlag{
				Name:    "icon",
				Usage:   "Icon file path",
				Aliases: []string{"ic"},
			},
			&cli.BoolFlag{
				Name:    "no-console",
				Usage:   "Hide console window",
				Aliases: []string{"nc"},
			},
			&cli.BoolFlag{
				Name:    "onefile",
				Usage:   "Create single executable file",
				Aliases: []string{"of"},
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			appName := cmd.String("name")
			icon := cmd.String("icon")
			noConsole := cmd.Bool("no-console")
			oneFile := cmd.Bool("onefile")

			return buildApplication(appName, icon, noConsole, oneFile)
		},
	}
}

func buildApplication(appName, icon string, noConsole, oneFile bool) error {
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

	if appName == "" {
		appName = cfg.Build.AppName
		if appName == "" {
			appName = filepath.Base(projectDir)
		}
	}

	if !oneFile && cfg.Build.OneFile {
		oneFile = true
	}
	if !noConsole && cfg.Build.NoConsole {
		noConsole = true
	}

	if !executor.CommandExists("uv") {
		return fmt.Errorf("uv is not installed. Please install it first")
	}

	logger.Info("Building application: %s", appName)

	logger.Info("Installing build dependencies...")
	ctx := context.Background()
	if err := executor.RunCommand(ctx, projectDir, "uv", "sync", "--extra", "build"); err != nil {
		return fmt.Errorf("failed to install build dependencies: %v", err)
	}

	distDir := filepath.Join(projectDir, "dist")
	buildDir := filepath.Join(projectDir, "build")

	if executor.DirExists(distDir) {
		logger.Info("Cleaning previous build...")
		os.RemoveAll(distDir)
	}
	if executor.DirExists(buildDir) {
		os.RemoveAll(buildDir)
	}

	webDir := filepath.Join(projectDir, "web")
	if executor.DirExists(webDir) {
		logger.Info("Building web assets...")
		if err := buildWebAssets(webDir, cfg.Manager); err != nil {
			return fmt.Errorf("failed to build web assets: %v", err)
		}
	}

	args := []string{"run", "pyinstaller", "--clean"}

	args = append(args, "--name", appName)
	args = append(args, "--noconfirm")

	if noConsole {
		args = append(args, "--noconsole")
	}

	args = append(args, "--paths", projectDir)

	if oneFile {
		args = append(args, "--onefile")
	} else {
		args = append(args, "--onedir")
	}

	if icon != "" {
		if !executor.FileExists(icon) {
			return fmt.Errorf("icon file not found: %s", icon)
		}
		args = append(args, "--icon", icon)
	}

	if executor.DirExists(webDir) {
		args = append(args, "--add-data", ".distweb;.distweb")
	}

	mainPath := "main.py"
	mainPath = ".\\main.py"
	args = append(args, mainPath)

	logger.Info("Running PyInstaller with args: %s", strings.Join(args, " "))

	if err := executor.RunCommand(ctx, projectDir, "uv", args...); err != nil {
		return fmt.Errorf("failed to build application: %v", err)
	}

	logger.Success("Build completed successfully!")
	logger.Info("Output directory: %s", distDir)

	if oneFile {
		exePath := filepath.Join(distDir, appName+".exe")
		if executor.FileExists(exePath) {
			logger.Info("Executable: %s", exePath)
		}
	} else {
		appDir := filepath.Join(distDir, appName)
		if executor.DirExists(appDir) {
			logger.Info("Application directory: %s", appDir)
		}
	}

	return nil
}

func buildWebAssets(webDir, manager string) error {
	executor := utils.NewExecutor()
	ctx := context.Background()

	if !executor.FileExists(filepath.Join(webDir, "package.json")) {
		return fmt.Errorf("package.json not found in web directory")
	}

	if !executor.DirExists(filepath.Join(webDir, "node_modules")) {
		if err := installWebDependencies(webDir, manager); err != nil {
			return fmt.Errorf("failed to install web dependencies: %v", err)
		}
	}

	var args []string
	switch manager {
	case "npm", "yarn", "pnpm", "bun":
		args = []string{"run", "build"}
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	return executor.RunCommand(ctx, webDir, manager, args...)
}
