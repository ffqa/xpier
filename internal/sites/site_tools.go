package sites

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

// ResolveSite returns the site for the current directory (by basename), or an
// explicit --site name.
func ResolveSite(sites *store.Sites, siteName string) (string, store.Site, error) {
	if siteName != "" {
		if s, ok := sites.Sites[siteName]; ok {
			return siteName, s, nil
		}
		return "", store.Site{}, fmt.Errorf("site %s is not linked", siteName)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", store.Site{}, err
	}
	name := filepath.Base(cwd)
	if s, ok := sites.Sites[name]; ok {
		return name, s, nil
	}
	// A site linked under an explicit name (e.g. `xpier link smoketest`) has a
	// different key than the directory basename; match by resolved path.
	absCwd, err := filepath.Abs(cwd)
	if err == nil {
		for n, s := range sites.Sites {
			if p, err := filepath.Abs(s.Path); err == nil && p == absCwd {
				return n, s, nil
			}
		}
	}
	return "", store.Site{}, fmt.Errorf("no site for current directory; link it with `xpier link` or pass --site")
}

// SitePHPBin resolves the php binary for a site (pinned version, manifest, or default).
func SitePHPBin(site store.Site) (string, string, error) {
	ver := site.PHP
	if ver == "" {
		ver = nginx.DefaultPhpVersion()
	}
	bin := filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "bin", "php")
	if !store.FileExists(bin) {
		return "", "", fmt.Errorf("php@%s not found at %s (run `brew install shivammathur/php/php@%s`)", ver, bin, ver)
	}
	return bin, ver, nil
}

func CmdIsolate(args []string) error {
	fs := flag.NewFlagSet("isolate", flag.ExitOnError)
	siteFlag := fs.String("site", "", "site name (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier isolate <php-version> [--site x]")
	}
	ver := fs.Arg(0)
	if !store.SafePhpRe.MatchString(ver) {
		return fmt.Errorf("invalid php version %q", ver)
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name, site, err := ResolveSite(sites, *siteFlag)
	if err != nil {
		return err
	}
	site.PHP = ver
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
	fmt.Printf("%s isolated to php %s\n", store.SiteDomain(sites, name), ver)
	return nil
}

func CmdUnisolate(args []string) error {
	fs := flag.NewFlagSet("unisolate", flag.ExitOnError)
	siteFlag := fs.String("site", "", "site name (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name, site, err := ResolveSite(sites, *siteFlag)
	if err != nil {
		return err
	}
	site.PHP = ""
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
	fmt.Printf("%s unisolated (php %s)\n", store.SiteDomain(sites, name), nginx.DefaultPhpVersion())
	return nil
}

func CmdIsolated(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	names := make([]string, 0)
	for name, site := range sites.Sites {
		if site.PHP != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Println("no isolated sites")
		return nil
	}
	for _, name := range names {
		fmt.Printf("  %-30s php %s\n", store.SiteDomain(sites, name), sites.Sites[name].PHP)
	}
	return nil
}

// ExtractSiteFlag pulls --site=x / --site x out of arbitrary command args,
// leaving everything else untouched for passthrough (php -r, etc).
func ExtractSiteFlag(args []string) (site string, rest []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--site" && i+1 < len(args):
			site = args[i+1]
			i++
		case strings.HasPrefix(a, "--site="):
			site = strings.TrimPrefix(a, "--site=")
		default:
			rest = append(rest, a)
		}
	}
	return site, rest
}

// runSitePHP proxies a command to the site's PHP binary (like `herd php`).
func runSitePHP(args []string, extra []string) error {
	siteName, passthrough := ExtractSiteFlag(args)
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := ResolveSite(sites, siteName)
	if err != nil {
		return err
	}
	bin, ver, err := SitePHPBin(site)
	if err != nil {
		return err
	}
	cmdArgs := append(append([]string{}, extra...), passthrough...)
	cmd := exec.Command(bin, cmdArgs...)
	cmd.Dir = site.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("php@%s exited: %w", ver, err)
	}
	return nil
}

func CmdSitePHPProxy(args []string) error { return runSitePHP(args, nil) }

func CmdSiteComposer(args []string) error {
	siteName, passthrough := ExtractSiteFlag(args)
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := ResolveSite(sites, siteName)
	if err != nil {
		return err
	}
	bin, ver, err := SitePHPBin(site)
	if err != nil {
		return err
	}
	composer, err := exec.LookPath("composer")
	if err != nil {
		return fmt.Errorf("composer not on PATH")
	}
	cmd := exec.Command(bin, append([]string{composer}, passthrough...)...)
	cmd.Dir = site.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("composer (php@%s) exited: %w", ver, err)
	}
	return nil
}

func CmdSiteDebug(args []string) error {
	return runSitePHP(args, []string{"-d", "xdebug.mode=debug", "-d", "xdebug.start_with_request=yes"})
}

func CmdSiteCoverage(args []string) error {
	return runSitePHP(args, []string{"-d", "xdebug.mode=coverage"})
}

func CmdOpen(args []string) error {
	fs := flag.NewFlagSet("open", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name := fs.Arg(0)
	if name == "" {
		cwd, _ := os.Getwd()
		name = filepath.Base(cwd)
	}
	url := "http://" + store.SiteDomain(sites, name)
	return store.RunOutErr("open", url)
}

func CmdEdit(args []string) error {
	fs := flag.NewFlagSet("edit", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name := fs.Arg(0)
	if name == "" {
		cwd, _ := os.Getwd()
		name = filepath.Base(cwd)
	}
	site, ok := sites.Sites[name]
	if !ok {
		return fmt.Errorf("site %s is not linked", name)
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if store.FileExists("/usr/local/bin/code") {
			editor = "/usr/local/bin/code"
		} else if store.FileExists("/Applications/Visual Studio Code.app") {
			editor = "open -a \"Visual Studio Code\""
		} else {
			return fmt.Errorf("no editor found; set $EDITOR")
		}
	}
	return store.RunOutErr(editor, site.Path)
}

func CmdSiteInformation(args []string) error {
	fs := flag.NewFlagSet("site-information", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name := fs.Arg(0)
	if name == "" {
		cwd, _ := os.Getwd()
		name = filepath.Base(cwd)
	}
	site, ok := sites.Sites[name]
	if !ok {
		return fmt.Errorf("site %s is not linked", name)
	}
	php := site.PHP
	if php == "" {
		php = nginx.DefaultPhpVersion()
	}
	fmt.Printf("domain:  %s\n", store.SiteDomain(sites, name))
	fmt.Printf("path:    %s\n", site.Path)
	fmt.Printf("driver:  %s\n", site.Driver)
	fmt.Printf("php:     %s\n", php)
	fmt.Printf("root:    %s\n", siteRoot(site))
	return nil
}

func CmdTLD(args []string) error {
	fs := flag.NewFlagSet("tld", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if fs.NArg() > 0 {
		tld := strings.TrimPrefix(fs.Arg(0), ".")
		if !store.SafeSiteNameRe.MatchString(tld) {
			return fmt.Errorf("invalid tld %q", tld)
		}
		sites.TLD = tld
		if err := sites.Save(); err != nil {
			return err
		}
		if err := store.WriteDnsmasqConfig(tld); err != nil {
			return err
		}
		fmt.Printf("tld set to .%s (dnsmasq config updated; run `sudo xpier install` to apply DNS)\n", tld)
		return nil
	}
	fmt.Println("." + sites.TLD)
	return nil
}

func CmdLoopback(args []string) error {
	fs := flag.NewFlagSet("loopback", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		lb := fs.Arg(0)
		sites, err := store.LoadSites()
		if err != nil {
			return err
		}
		_ = lb
		fmt.Println("loopback is fixed at 127.0.0.1 in xpier")
		return sites.Save()
	}
	fmt.Println("127.0.0.1")
	return nil
}

func CmdLinks(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(sites.Sites))
	for name := range sites.Sites {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %-30s %s\n", store.SiteDomain(sites, name), sites.Sites[name].Path)
	}
	return nil
}

func CmdParked(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	for _, p := range sites.Parked {
		fmt.Printf("  %s\n", p)
	}
	return nil
}
