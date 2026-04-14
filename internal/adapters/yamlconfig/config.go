package yamlconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

type yamlConfig struct {
	Collections []string `yaml:"collections"`
}

type configLoader struct {
	path string
}

func New(path string) core.ConfigLoader {
	return &configLoader{path: path}
}

func (l *configLoader) Load() (core.Config, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return core.Config{}, fmt.Errorf("config file not found at %s — create it with your collection paths", l.path)
		}
		return core.Config{}, fmt.Errorf("failed to read config at %s: %w", l.path, err)
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return core.Config{}, fmt.Errorf("failed to parse config at %s: %w", l.path, err)
	}

	paths := make([]string, len(yc.Collections))
	for i, p := range yc.Collections {
		expanded, err := expandPath(p)
		if err != nil {
			return core.Config{}, fmt.Errorf("failed to expand path %q: %w", p, err)
		}
		paths[i] = expanded
	}

	return core.Config{CollectionPaths: paths}, nil
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bruwrapper.yaml")
}

func expandPath(p string) (string, error) {
	if len(p) >= 2 && p[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}
