package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// Config struct holds the configuration settings for the application.
type Config struct {
	Keywords          []string `yaml:"keywords"`
	ArtifactsPath     string   `yaml:"artifacts_path"`
	DefaultSourcePath string   `yaml:"default_source_path"`
}

func newConfig() Config {
	cfg := Config{
		Keywords:          []string{},
		ArtifactsPath:     "artifacts",
		DefaultSourcePath: "",
	}

	return cfg
}

// Load reads the configuration from the specified YAML file and returns a Config struct.
// If the file does not exist, it returns a default configuration, otherwise it unmarshals the YAML content into the
// Config struct or returns an error if the file cannot be read or the YAML is invalid.
func Load(path string) (Config, error) {
	cfg := newConfig()

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal yaml in %s: %w", path, err)
	}

	return cfg, nil
}
