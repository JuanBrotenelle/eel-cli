package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Executor struct {
	logger *Logger
}

func NewExecutor() *Executor {
	return &Executor{
		logger: NewLogger(),
	}
}

func (e *Executor) RunCommand(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	e.logger.Info("Running: %s %s in %s", name, strings.Join(args, " "), dir)

	return cmd.Run()
}

func (e *Executor) RunCommandSilent(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	return cmd.Run()
}

func (e *Executor) CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (e *Executor) GetWorkingDir() (string, error) {
	return os.Getwd()
}

func (e *Executor) ChangeDir(dir string) error {
	return os.Chdir(dir)
}

func (e *Executor) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (e *Executor) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (e *Executor) DirExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}

func (e *Executor) JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}
