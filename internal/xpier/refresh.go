package xpier

import (
	"fmt"
	"xpier/internal/nginx"
	"xpier/internal/sites"
	"xpier/internal/store"
)

// cmdRefresh regenerates all site/app nginx configs with current paths and
// reloads nginx. Needed after the tool's home directory is moved/renamed.
func cmdRefresh(args []string) error {
	s, err := store.LoadSites()
	if err != nil {
		return err
	}
	sites.SyncParked(s)
	if err := nginx.WriteAllSiteConfigs(s); err != nil {
		return err
	}
	if err := nginx.WriteDefaultSiteConfig(); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		return err
	}
	fmt.Println("nginx configs regenerated and reloaded")
	return nil
}
