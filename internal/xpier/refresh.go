package xpier

import "fmt"

// cmdRefresh regenerates all site/app nginx configs with current paths and
// reloads nginx. Needed after the tool's home directory is moved/renamed.
func cmdRefresh(args []string) error {
	sites, err := loadSites()
	if err != nil {
		return err
	}
	syncParked(sites)
	if err := writeAllSiteConfigs(sites); err != nil {
		return err
	}
	if err := writeDefaultSiteConfig(); err != nil {
		return err
	}
	if err := nginxReload(); err != nil {
		return err
	}
	fmt.Println("nginx configs regenerated and reloaded")
	return nil
}
