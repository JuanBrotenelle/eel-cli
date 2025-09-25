package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eel-cli/internal/config"
	"eel-cli/internal/template"
	"eel-cli/pkg/utils"

	"github.com/urfave/cli/v3"
)

func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new Eel project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "manager",
				Usage:   "Package manager (npm, yarn, pnpm, bun)",
				Aliases: []string{"pm"},
			},
			&cli.StringFlag{
				Name:    "init",
				Usage:   "Command to run in web directory after creation",
				Aliases: []string{"i"},
			},
		},
		Action: func(c context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) == 0 {
				return fmt.Errorf("project name is required")
			}

			projectName := args[0]
			manager := cmd.String("manager")
			initRaw := parseMultiWordFlag(os.Args, "--init", "-i")
			templateName := strings.Fields(initRaw)
			selectedTemplate := ""
			if len(templateName) > 0 {
				selectedTemplate = templateName[0]
			}

			if manager == "" {
				var err error
				manager, err = promptForManager()
				if err != nil {
					return err
				}
			}

			if !isValidManager(manager) {
				return fmt.Errorf("invalid package manager: %s. Supported: npm, yarn, pnpm, bun", manager)
			}

			return createProject(projectName, manager, selectedTemplate)
		},
	}
}

func promptForManager() (string, error) {
	fmt.Print("🐍 eel-cli: Select package manager (npm/yarn/pnpm/bun): ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	if !isValidManager(input) {
		return "", fmt.Errorf("invalid package manager: %s", input)
	}

	return input, nil
}

func parseMultiWordFlag(argv []string, long string, short string) string {
	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		if tok == long || tok == short {
			if strings.Contains(tok, "=") {
				parts := strings.SplitN(tok, "=", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
			var collected []string
			for j := i + 1; j < len(argv); j++ {
				if strings.HasPrefix(argv[j], "-") {
					break
				}
				collected = append(collected, argv[j])
			}
			return strings.TrimSpace(strings.Join(collected, " "))
		}
		if strings.HasPrefix(tok, long+"=") {
			return strings.TrimSpace(strings.TrimPrefix(tok, long+"="))
		}
	}
	return ""
}

func isValidManager(manager string) bool {
	validManagers := []string{"npm", "yarn", "pnpm", "bun"}
	for _, valid := range validManagers {
		if manager == valid {
			return true
		}
	}
	return false
}

func createProject(projectName, manager, templateName string) error {
	logger := utils.NewLogger()
	executor := utils.NewExecutor()

	if executor.DirExists(projectName) {
		return fmt.Errorf("directory %s already exists", projectName)
	}

	logger.Info("Creating project: %s", projectName)

	if err := executor.CreateDir(projectName); err != nil {
		return fmt.Errorf("failed to create project directory: %v", err)
	}

	webDir := executor.JoinPath(projectName, "web")
	if err := scaffoldWebWithVite(projectName, manager, templateName); err != nil {
		return fmt.Errorf("failed to scaffold web with Vite: %v", err)
	}

	if err := template.CopyTemplateFiles(projectName); err != nil {
		return fmt.Errorf("failed to copy template files: %v", err)
	}

	if err := ensureViteBuildConfig(webDir); err != nil {
		logger.Warning("Could not update vite config: %v", err)
	}

	cfg := &config.Config{
		Manager: manager,
		Dev: config.DevConfig{
			Mode: "url",
		},
		Build: config.BuildConfig{
			AppName:   projectName,
			Icon:      "",
			NoConsole: true,
			OneFile:   true,
		},
	}

	if err := config.SaveConfig(projectName, cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	logger.Success("Project %s created successfully!", projectName)
	logger.Info("Next steps:")
	logger.Info("  cd %s", projectName)
	logger.Info("  eel install")
	logger.Info("  eel dev")

	return nil
}

func scaffoldWebWithVite(projectDir, manager, templateName string) error {
	executor := utils.NewExecutor()
	ctx := context.Background()

	tmpl := templateName
	if tmpl == "" {
		tmpl = "vanilla"
	}

	args := []string{"create", "vite", "web", "--template", tmpl, "--no-rolldown", "--interactive", "--no-immediate"}

	return executor.RunCommand(ctx, projectDir, manager, args...)
}

func ensureViteBuildConfig(webDir string) error {
	executor := utils.NewExecutor()
	candidates := []string{
		filepath.Join(webDir, "vite.config.ts"),
		filepath.Join(webDir, "vite.config.js"),
		filepath.Join(webDir, "vite.config.mts"),
		filepath.Join(webDir, "vite.config.mjs"),
		filepath.Join(webDir, "vite.config.cts"),
		filepath.Join(webDir, "vite.config.cjs"),
	}

	var cfgPath string
	for _, p := range candidates {
		if executor.FileExists(p) {
			cfgPath = p
			break
		}
	}
	if cfgPath == "" {
		return fmt.Errorf("vite config not found in %s", webDir)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	content := string(data)

	if strings.Contains(content, "build:") && strings.Contains(content, "outDir:") {
		return nil
	}

	insertBlock := "build: {\n    outDir: '../.distweb',\n    emptyOutDir: true,\n  },"

	idx := strings.Index(content, "defineConfig(")
	if idx >= 0 {
		braceIdx := strings.Index(content[idx:], "{")
		if braceIdx >= 0 {
			pos := idx + braceIdx + 1
			updated := content[:pos] + "\n  " + insertBlock + content[pos:]
			return os.WriteFile(cfgPath, []byte(updated), 0644)
		}
	}

	firstBrace := strings.Index(content, "{")
	if firstBrace >= 0 {
		updated := content[:firstBrace+1] + "\n  " + insertBlock + content[firstBrace+1:]
		return os.WriteFile(cfgPath, []byte(updated), 0644)
	}

	return fmt.Errorf("could not inject build config into %s", cfgPath)
}
