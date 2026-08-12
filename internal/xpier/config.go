package xpier

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manifest is the project-declared environment. It is optional: xpier
// auto-detects runtime and PHP, and the manifest only exists to pin or
// declare things that cannot be detected. Empty fields are not stored.
type Manifest struct {
	PHP        string            `yaml:"php,omitempty"`
	Runtime    string            `yaml:"runtime,omitempty"`
	Extensions map[string]string `yaml:"extensions,omitempty"`
	Services   []string          `yaml:"services,omitempty"`
	Apps       map[string]App    `yaml:"apps,omitempty"`
}

func DefaultManifest() *Manifest {
	return &Manifest{
		Runtime: "fpm",
	}
}

func (m *Manifest) save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func cmdInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	php := fs.String("php", "", "pin PHP version, e.g. 8.2")
	runtime := fs.String("runtime", "", "runtime (fpm | hyperf | swoole | frankenphp)")
	local := fs.Bool("local", false, "write xpier.yaml into the current directory instead of ~/.xpier (commit it to git if you want it versioned)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 && fs.Arg(0) == "." {
		*local = true
	}
	if *runtime != "" {
		switch *runtime {
		case "fpm", "hyperf", "swoole", "frankenphp":
		default:
			return fmt.Errorf("invalid runtime %q (fpm | hyperf | swoole | frankenphp)", *runtime)
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var manifestPath string
	if *local {
		manifestPath = filepath.Join(cwd, manifestName)
	} else {
		manifestPath, _ = projectPaths(cwd)
		if err := ensureProjectDir(cwd); err != nil {
			return err
		}
	}
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("%s already exists", manifestPath)
	}
	m := DefaultManifest()
	if *php != "" {
		m.PHP = *php
	}
	if *runtime != "" {
		m.Runtime = *runtime
	}
	if err := m.save(manifestPath); err != nil {
		return err
	}
	fmt.Printf("created %s (php %s, runtime %s)\n", manifestPath, m.PHP, m.Runtime)
	fmt.Println("manifest is optional: xpier auto-detects runtime and PHP; the manifest only pins them.")
	return nil
}
