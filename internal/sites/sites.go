package sites

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/service"
	"xpier/internal/store"
)

// DetectDriver classifies a project directory.
//
//	laravel: public/index.php (PHP-FPM)
//	hyperf:  bin/hyperf.php (reverse-proxy to the runtime port)
//	spa:     dist/index.html (static, dist as document root)
//	static:  anything else
func DetectDriver(dir string) string {
	if store.FileExists(filepath.Join(dir, "bin", "hyperf.php")) {
		return "hyperf"
	}
	if store.FileExists(filepath.Join(dir, "public", "index.php")) {
		return "laravel"
	}
	if store.FileExists(filepath.Join(dir, "dist", "index.html")) {
		return "spa"
	}
	return "static"
}

// siteRoot returns the nginx document root for a site.
func siteRoot(site store.Site) string {
	switch site.Driver {
	case "laravel":
		return filepath.Join(site.Path, "public")
	case "spa":
		return filepath.Join(site.Path, "dist")
	}
	return site.Path
}

// nginx.HyperfPort returns the proxy port for a hyperf site, reading
// config/autoload/server.php when possible.
// CmdPark registers directories whose subdirectories become sites.
func CmdPark(args []string) error {
	fs := flag.NewFlagSet("park", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier park <directory> [directory ...]")
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	for _, raw := range fs.Args() {
		abs, err := filepath.Abs(raw)
		if err != nil {
			return err
		}
		if _, err := os.Stat(abs); err != nil {
			return fmt.Errorf("park %s: %w", abs, err)
		}
		if !containsString(sites.Parked, abs) {
			sites.Parked = append(sites.Parked, abs)
		}
		fmt.Printf("parked %s\n", abs)
	}
	SyncParked(sites)
	if err := sites.Save(); err != nil {
		return err
	}
	for _, name := range store.SortedKeys(sites.Sites) {
		fmt.Printf("  site %s -> %s (%s)\n", store.SiteDomain(sites, name), sites.Sites[name].Path, sites.Sites[name].Driver)
	}
	if err := nginx.WriteAllSiteConfigs(sites); err != nil {
		return err
	}
	return nginx.NginxReload()
}

// SyncParked auto-registers immediate subdirectories of parked paths.
func SyncParked(sites *store.Sites) {
	for _, dir := range sites.Parked {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			name := e.Name()
			if _, exists := sites.Sites[name]; exists {
				continue
			}
			path := filepath.Join(dir, name)
			sites.Sites[name] = store.Site{Path: path, Driver: DetectDriver(path)}
		}
	}
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// nginx.WriteAllSiteConfigs regenerates nginx configs for every registered site.
func CmdLink(args []string) error {
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	name := fs.String("name", "", "site name (default: directory name)")
	php := fs.String("php", "", "pin PHP version for this site")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	siteName := *name
	if siteName == "" && fs.NArg() > 0 {
		siteName = fs.Arg(0) // positional: xpier link abc.def
	}
	if siteName == "" {
		siteName = filepath.Base(cwd)
	}
	if !store.SafeSiteNameRe.MatchString(siteName) {
		return fmt.Errorf("invalid site name %q (use [a-z0-9._-])", siteName)
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if _, exists := sites.Sites[siteName]; exists {
		return fmt.Errorf("site %s already linked (unlink first)", siteName)
	}
	site := store.Site{Path: cwd, Driver: DetectDriver(cwd)}
	if *php != "" {
		site.PHP = *php
	} else if m, err := nginx.LoadManifestFrom(cwd); err == nil && m.PHP != "" {
		site.PHP = m.PHP
	}
	sites.Sites[siteName] = site
	if err := sites.Save(); err != nil {
		return err
	}
	if err := nginx.WriteSiteNginxConfig(sites, siteName); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v (run `sudo xpier install` first?)\n", err)
	}
	fmt.Printf("linked %s -> %s (driver %s, php %s)\n", store.SiteDomain(sites, siteName), cwd, site.Driver, site.PHP)
	return nil
}

func CmdUnlink(args []string) error {
	fs := flag.NewFlagSet("unlink", flag.ExitOnError)
	name := fs.String("name", "", "site name (default: directory name)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	siteName := *name
	if siteName == "" && fs.NArg() > 0 {
		siteName = fs.Arg(0)
	}
	if siteName == "" {
		siteName = filepath.Base(cwd)
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if _, exists := sites.Sites[siteName]; !exists {
		return fmt.Errorf("site %s is not linked", siteName)
	}
	delete(sites.Sites, siteName)
	if err := sites.Save(); err != nil {
		return err
	}
	if err := nginx.RemoveSiteNginxConfig(siteName); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	fmt.Printf("unlinked %s\n", store.SiteDomain(sites, siteName))
	return nil
}

func CmdSitePHP(args []string) error {
	fs := flag.NewFlagSet("site:php", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 || fs.NArg() > 2 {
		return fmt.Errorf("usage: xpier site:php <site> [version]")
	}
	siteName := fs.Arg(0)
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	site, exists := sites.Sites[siteName]
	if !exists {
		return fmt.Errorf("site %s is not linked", siteName)
	}
	if fs.NArg() == 2 {
		site.PHP = fs.Arg(1)
		sites.Sites[siteName] = site
		if err := sites.Save(); err != nil {
			return err
		}
		if err := nginx.WriteSiteNginxConfig(sites, siteName); err != nil {
			return err
		}
		if err := nginx.NginxReload(); err != nil {
			fmt.Printf("[warn] nginx reload failed: %v\n", err)
		}
		fmt.Printf("%s -> php %s\n", store.SiteDomain(sites, siteName), site.PHP)
		return nil
	}
	fmt.Printf("%s -> php %s\n", store.SiteDomain(sites, siteName), site.PHP)
	return nil
}

func CmdSites(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	SyncParked(sites)
	if err := sites.Save(); err != nil {
		return err
	}
	nginxUp := false
	if b, _ := store.PortBusy("80"); b {
		nginxUp = true
	}
	dnsUp := false
	if b, _ := store.UDPBusy("53"); b {
		dnsUp = true
	}
	fmt.Printf("nginx: %s | dnsmasq: %s | %d site(s)\n", store.UpDown(nginxUp), store.UpDown(dnsUp), len(sites.Sites))
	names := make([]string, 0, len(sites.Sites))
	for name := range sites.Sites {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		site := sites.Sites[name]
		php := site.PHP
		if php == "" {
			php = nginx.DefaultPhpVersion()
		}
		up := service.FpmRunning(php)
		fmt.Printf("  %-30s driver=%-7s php=%-4s fpm=%s path=%s\n",
			store.SiteDomain(sites, name), site.Driver, php, store.UpDown(up), site.Path)
	}
	return nil
}
