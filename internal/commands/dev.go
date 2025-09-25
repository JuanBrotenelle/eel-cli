package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"eel-cli/internal/config"
	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func DevCommand() *cli.Command {
	return &cli.Command{
		Name:  "dev",
		Usage: "Start development server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "mode",
				Usage:   "Development mode (watch, url)",
				Aliases: []string{"m"},
				Value:   "url",
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			mode := cmd.String("mode")

			if mode != "watch" && mode != "url" {
				return fmt.Errorf("invalid mode: %s. Supported modes: watch, url", mode)
			}

			return startDevServer(mode)
		},
	}
}

func startDevServer(mode string) error {
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

	webDir := filepath.Join(projectDir, "web")
	if !executor.DirExists(webDir) {
		return fmt.Errorf("web directory not found")
	}

	manager := cfg.Manager
	if manager == "" {
		manager = detectPackageManager()
	}

	logger.Info("Starting development server in %s mode", mode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down development server...")
		cancel()
	}()

	if mode == "url" {
		return startURLMode(ctx, projectDir, webDir, manager, logger)
	} else {
		return startWatchMode(ctx, projectDir, webDir, manager, logger)
	}
}

func startURLMode(ctx context.Context, projectDir, webDir, manager string, logger *utils.Logger) error {
	// Check if node_modules exists
	if !utils.NewExecutor().DirExists(filepath.Join(webDir, "node_modules")) {
		logger.Info("Installing web dependencies...")
		if err := installWebDependencies(webDir, manager); err != nil {
			return fmt.Errorf("failed to install web dependencies: %v", err)
		}
	}

	vitePort := 5173
	viteHost := "localhost"
	viteURL := fmt.Sprintf("http://%s:%d", viteHost, vitePort)

	logger.Info("Starting Vite dev server on %s", viteURL)

	var viteArgs []string
	switch manager {
	case "bun":
		viteArgs = []string{"run", "dev", "--port", strconv.Itoa(vitePort), "--host", viteHost}
	case "npm", "yarn", "pnpm":
		viteArgs = []string{"run", "dev", "--", "--port", strconv.Itoa(vitePort), "--host", viteHost}
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	viteCmd := exec.CommandContext(ctx, manager, viteArgs...)
	viteCmd.Dir = webDir
	viteCmd.Stdout = os.Stdout
	viteCmd.Stderr = os.Stderr

	if err := viteCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Vite: %v", err)
	}

	if err := waitForURL(viteURL, 15*time.Second); err != nil {
		viteCmd.Process.Kill()
		return fmt.Errorf("Vite server failed to start: %v", err)
	}

	logger.Success("Vite dev server is ready at %s", viteURL)

	os.Setenv("VITE_DEV_SERVER_URL", viteURL)

	logger.Info("Starting Eel application...")
	eelCmd := exec.CommandContext(ctx, "uv", "run", "python", "main.py")
	eelCmd.Dir = projectDir
	eelCmd.Stdout = os.Stdout
	eelCmd.Stderr = os.Stderr

	if err := eelCmd.Start(); err != nil {
		viteCmd.Process.Kill()
		return fmt.Errorf("failed to start Eel: %v", err)
	}

	done := make(chan error, 2)
	go func() {
		done <- viteCmd.Wait()
	}()
	go func() {
		done <- eelCmd.Wait()
	}()

	err := <-done
	if err != nil {
		logger.Warning("Process exited with error: %v", err)
	}

	viteCmd.Process.Kill()
	eelCmd.Process.Kill()

	return nil
}

func startWatchMode(ctx context.Context, projectDir, webDir, manager string, logger *utils.Logger) error {
	if !utils.NewExecutor().DirExists(filepath.Join(webDir, "node_modules")) {
		logger.Info("Installing web dependencies...")
		if err := installWebDependencies(webDir, manager); err != nil {
			return fmt.Errorf("failed to install web dependencies: %v", err)
		}
	}

	logger.Info("Starting build watch...")

	var watchArgs []string
	switch manager {
	case "bun":
		watchArgs = []string{"run", "build", "--watch"}
	case "npm", "yarn", "pnpm":
		watchArgs = []string{"run", "build", "--", "--watch"}
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	watchCmd := exec.CommandContext(ctx, manager, watchArgs...)
	watchCmd.Dir = webDir
	watchCmd.Stdout = os.Stdout
	watchCmd.Stderr = os.Stderr

	if err := watchCmd.Start(); err != nil {
		return fmt.Errorf("failed to start build watch: %v", err)
	}

	logger.Info("Starting Eel application...")
	eelCmd := exec.CommandContext(ctx, "uv", "run", "python", "main.py")
	eelCmd.Dir = projectDir
	eelCmd.Stdout = os.Stdout
	eelCmd.Stderr = os.Stderr

	if err := eelCmd.Start(); err != nil {
		watchCmd.Process.Kill()
		return fmt.Errorf("failed to start Eel: %v", err)
	}

	done := make(chan error, 2)
	go func() {
		done <- watchCmd.Wait()
	}()
	go func() {
		done <- eelCmd.Wait()
	}()

	err := <-done
	if err != nil {
		logger.Warning("Process exited with error: %v", err)
	}

	watchCmd.Process.Kill()
	eelCmd.Process.Kill()

	return nil
}

func waitForURL(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 500 {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(300 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s", url)
}
