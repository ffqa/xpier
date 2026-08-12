package sites

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"xpier/internal/nginx"
	"xpier/internal/service"
	"xpier/internal/store"
)

// CmdSitesUp starts php-fpm for every PHP version used by linked sites.
func CmdSitesUp(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	SyncParked(sites)
	if err := sites.Save(); err != nil {
		return err
	}
	if len(sites.Sites) == 0 {
		return fmt.Errorf("no linked sites; link one with `xpier link` or `xpier park <dir>` first")
	}
	if b, _ := store.PortBusy("80"); !b {
		fmt.Println("[warn] nginx is not listening on port 80; run `sudo xpier install` first")
	}
	versions := map[string]bool{}
	for _, site := range sites.Sites {
		ver := site.PHP
		if ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
		versions[ver] = true
	}
	for ver := range versions {
		if err := service.FpmUp(ver); err != nil {
			fmt.Printf("[warn] %v\n", err)
		}
	}
	return nil
}

// CmdSitesDown stops all managed php-fpm instances.
func CmdSitesDown(args []string) error {
	entries, err := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	if err != nil {
		return fmt.Errorf("no php-fpm state found")
	}
	stopped := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			if err := service.FpmDown(ver); err == nil {
				stopped++
			}
		}
	}
	if stopped == 0 {
		fmt.Println("no php-fpm running")
	}
	return nil
}
