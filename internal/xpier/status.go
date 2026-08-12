package xpier

import (
	"fmt"
	"os"
	"xpier/internal/apps"
	"xpier/internal/store"
)

func cmdStatus(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath, lockPath := store.ResolvePaths(cwd)
	m, mErr := store.LoadManifest(manifestPath)
	if mErr != nil {
		fmt.Printf("manifest: none (auto-detect mode; create pins with `xpier init --php 8.2`)\n")
	} else {
		fmt.Printf("manifest %s:\n", manifestPath)
		fmt.Printf("  php %s, runtime %s\n", m.PHP, m.Runtime)
		for _, ext := range store.SortedKeys(m.Extensions) {
			fmt.Printf("  ext %s: %s\n", ext, m.Extensions[ext])
		}
		for _, svc := range m.Services {
			fmt.Printf("  svc %s\n", svc)
		}
	}
	lock, err := store.LoadLock(lockPath)
	if err != nil {
		fmt.Printf("lock %s: absent (run `xpier sync --apply`)\n", lockPath)
	} else {
		fmt.Printf("lock %s:\n", lockPath)
		fmt.Printf("  php %s @ %s\n", lock.PHP.Version, lock.PHP.Path)
		for _, e := range lock.Extensions {
			flag := "not loaded"
			if e.Loaded {
				flag = "loaded"
			}
			fmt.Printf("  ext %s: %s (%s)\n", e.Name, e.Installed, flag)
		}
		for _, s := range lock.Services {
			fmt.Printf("  svc %s: running=%v\n", s.Name, s.Running)
		}
	}
	// store.App stack table (dev.yaml / xpier.yaml apps).
	if _, _, err := apps.LoadAppConfig(); err == nil {
		apps.CmdStatus(nil)
	}
	return nil
}
