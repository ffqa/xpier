package xpier

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"xpier/internal/store"
)

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
		manifestPath = filepath.Join(cwd, store.ManifestName)
	} else {
		manifestPath, _ = store.ProjectPaths(cwd)
		if err := store.EnsureProjectDir(cwd); err != nil {
			return err
		}
	}
	if _, err := os.Stat(manifestPath); err == nil {
		return fmt.Errorf("%s already exists", manifestPath)
	}
	m := store.DefaultManifest()
	if *php != "" {
		m.PHP = *php
	}
	if *runtime != "" {
		m.Runtime = *runtime
	}
	if err := m.Save(manifestPath); err != nil {
		return err
	}
	fmt.Printf("created %s (php %s, runtime %s)\n", manifestPath, m.PHP, m.Runtime)
	fmt.Println("manifest is optional: xpier auto-detects runtime and PHP; the manifest only pins them.")
	return nil
}
