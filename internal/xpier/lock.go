package xpier

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Lock records the exact environment provisioned for a project (xpier.lock).
type Lock struct {
	SchemaVersion int           `yaml:"schema_version"`
	GeneratedAt   string        `yaml:"generated_at"`
	PHP           PhpLock       `yaml:"php"`
	Extensions    []ExtLock     `yaml:"extensions"`
	Services      []ServiceLock `yaml:"services"`
}

type PhpLock struct {
	Version string `yaml:"version"`
	Path    string `yaml:"path"`
}

type ExtLock struct {
	Name       string `yaml:"name"`
	Constraint string `yaml:"constraint"`
	Installed  string `yaml:"installed"`
	Loaded     bool   `yaml:"loaded"`
}

type ServiceLock struct {
	Name    string `yaml:"name"`
	Running bool   `yaml:"running"`
}

func loadLock(path string) (*Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Lock
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *Lock) save(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
