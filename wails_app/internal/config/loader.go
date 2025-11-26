package config

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func LoadConfig(path string) (*Configuration, error) {
	cfg := new(Configuration)
	err := ini.MapTo(cfg, path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveConfig(path string, cfg *Configuration) error {
	f := ini.Empty()
	err := ini.ReflectFrom(f, cfg)
	if err != nil {
		return err
	}
	return f.SaveTo(path)
}

func GetDefaultConfigPath() string {
	// Assuming the config is in the parent directory of the executable or current working directory
	// For development, we might look in the parent of wails_app
	// Adjust logic as needed for production
	cwd, _ := os.Getwd()
	// Try to find config/config.ini in current or parent directories
	paths := []string{
		filepath.Join(cwd, "config", "config.ini"),
		filepath.Join(cwd, "..", "config", "config.ini"),
		filepath.Join(cwd, "..", "..", "config", "config.ini"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
