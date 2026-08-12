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
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v (run `sudo xpier install` first?)\n", err)
	}
	return nil
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

// CmdLink registers the current directory (or a named site) as a linked site.
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
	dnsUp := service.DnsmasqRunning()
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
		domain := store.SiteDomain(sites, name)
		if site.Secure != nil && !*site.Secure {
			domain += " (http-only)"
		}
		fmt.Printf("  %-30s driver=%-7s php=%-4s fpm=%s path=%s\n",
			domain, site.Driver, php, store.UpDown(up), site.Path)
	}
	return nil
}

// CmdPaths lists every registered project path: parked dirs plus site paths.
func CmdPaths(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if len(sites.Parked) == 0 && len(sites.Sites) == 0 {
		fmt.Println("no sites or parked paths yet (link or park something first)")
		return nil
	}
	fmt.Println("parked:")
	for _, p := range sites.Parked {
		fmt.Printf("  %s\n", p)
	}
	fmt.Println("sites:")
	for _, name := range store.SortedKeys(sites.Sites) {
		fmt.Printf("  %-28s %s\n", store.SiteDomain(sites, name), sites.Sites[name].Path)
	}
	return nil
}

// CmdWhich prints the PHP version used by the current directory's site
// (site pin, manifest, or the global default).
func CmdWhich(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, site, err := ResolveSite(sites, ""); err == nil && site.PHP != "" {
		fmt.Println(site.PHP)
		return nil
	}
	if m, err := nginx.LoadManifestFrom(cwd); err == nil && m.PHP != "" {
		fmt.Println(m.PHP)
		return nil
	}
	fmt.Println(nginx.DefaultPhpVersion())
	return nil
}

// CmdWhichPHP prints the full php binary path + version for the current site.
func CmdWhichPHP(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := ResolveSite(sites, "")
	if err != nil {
		return fmt.Errorf("%v (run from a linked site or pass --site)", err)
	}
	bin, ver, err := SitePHPBin(site)
	if err != nil {
		return err
	}
	out, err := store.RunOut(bin, "-v")
	if err != nil {
		return err
	}
	first := strings.SplitN(out, "\n", 2)[0]
	fmt.Printf("%s (%s)\n", bin, first)
	_ = ver
	return nil
}

// isHelpArg reports whether the first arg is a help flag (for commands
// that parse args manually instead of via the flag package).
func isHelpArg(args []string) bool {
	return len(args) > 0 && (args[0] == "-h" || args[0] == "--help")
}

// CmdUse pins the global default PHP version (`xpier use 8.3`).
func CmdUse(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier use [8.3]    show / set the global default PHP version")
		return nil
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if len(args) < 1 {
		if sites.DefaultPHP != "" {
			fmt.Printf("default PHP: %s\n", sites.DefaultPHP)
		} else {
			fmt.Printf("default PHP: %s (auto)\n", nginx.DefaultPhpVersion())
		}
		return nil
	}
	ver := args[0]
	if !store.SafePhpRe.MatchString(ver) {
		return fmt.Errorf("invalid php version %q", ver)
	}
	if !store.FileExists(phpBinForVer(ver)) {
		return fmt.Errorf("php@%s not installed (run `xpier php:install %s` or brew install shivammathur/php/php@%s)", ver, ver, ver)
	}
	sites.DefaultPHP = ver
	if err := sites.Save(); err != nil {
		return err
	}
	fmt.Printf("default PHP set to %s\n", ver)
	return nil
}

func phpBinForVer(ver string) string {
	return filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "bin", "php")
}

// CmdUnsecure serves a site over plain http only (removes its 443 block).
func CmdUnsecure(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier unsecure <site>    serve a site over http only")
		return nil
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name := ""
	if len(args) > 0 {
		name = args[0]
	} else if cwd, err := os.Getwd(); err == nil {
		name = filepath.Base(cwd)
	}
	site, ok := sites.Sites[name]
	if !ok {
		return fmt.Errorf("site %s is not linked", name)
	}
	f := false
	site.Secure = &f
	sites.Sites[name] = site
	if err := sites.Save(); err != nil {
		return err
	}
	if err := nginx.WriteSiteNginxConfig(sites, name); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	fmt.Printf("%s now serves over http only (run `xpier secure %s` to re-enable https)\n", store.SiteDomain(sites, name), name)
	return nil
}
