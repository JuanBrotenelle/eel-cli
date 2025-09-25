package template

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed files/*
var templateFiles embed.FS

func CopyTemplateFiles(projectDir string) error {
	return fs.WalkDir(templateFiles, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "files" {
			return nil
		}

		relPath := strings.TrimPrefix(path, "files/")
		targetPath := filepath.Join(projectDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := templateFiles.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, content, 0644)
	})
}
