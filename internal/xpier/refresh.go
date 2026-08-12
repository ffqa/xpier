package xpier

import (
	"fmt"
	"xpier/internal/apps"
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
	// Rebuild proxy and app configs too, so cert paths point at the current
	// home after a migration.
	proxies, err := store.LoadProxies()
	if err != nil {
		return err
	}
	for domain, upstream := range proxies {
		if err := writeProxyConf(domain, upstream); err != nil {
			return err
		}
	}
	if err := apps.RefreshNginxConfs(); err != nil {
		return err
	}
	if err := apps.MigrateStateLogs(); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		return err
	}
	fmt.Println("nginx configs regenerated and reloaded")
	return nil
}
