package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Manager string      `json:"manager"`
	Dev     DevConfig   `json:"dev"`
	Build   BuildConfig `json:"build"`
}

type DevConfig struct {
	Mode string `json:"mode"`
}

type BuildConfig struct {
	AppName   string `json:"appName"`
	Icon      string `json:"icon"`
	NoConsole bool   `json:"noConsole"`
	OneFile   bool   `json:"oneFile"`
}

func LoadConfig(projectDir string) (*Config, error) {
	configPath := filepath.Join(projectDir, "eel.cli.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Manager: "",
			Dev: DevConfig{
				Mode: "",
			},
			Build: BuildConfig{
				AppName:   "",
				Icon:      "",
				NoConsole: true,
				OneFile:   true,
			},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(projectDir string, config *Config) error {
	configPath := filepath.Join(projectDir, "eel.cli.json")

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
