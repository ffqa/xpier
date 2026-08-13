package xpier

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/service"
	"xpier/internal/share"
	sitepkg "xpier/internal/sites"
	"xpier/internal/store"
)

// cmdTinker runs Laravel Tinker with the site's PHP.
func cmdTinker(args []string) error {
	siteName, passthrough := sitepkg.ExtractSiteFlag(args)
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := sitepkg.ResolveSite(sites, siteName)
	if err != nil {
		return err
	}
	bin, ver, err := sitepkg.SitePHPBin(site)
	if err != nil {
		return err
	}
	// Auto-detect the app type: Hyperf (bin/hyperf.php) or Laravel (artisan).
	var script string
	if store.FileExists(filepath.Join(site.Path, "bin", "hyperf.php")) {
		script = "bin/hyperf.php"
	} else if store.FileExists(filepath.Join(site.Path, "artisan")) {
		script = "artisan"
	} else {
		return fmt.Errorf("no tinker entry found in %s (neither bin/hyperf.php nor artisan)", site.Path)
	}
	cmd := exec.Command(bin, append([]string{script, "tinker"}, passthrough...)...)
	cmd.Dir = site.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tinker (php@%s) exited: %w", ver, err)
	}
	return nil
}

// cmdDirectoryListing toggles nginx autoindex in the main config.
func cmdDirectoryListing(args []string) error {
	on := false
	if len(args) > 0 {
		switch args[0] {
		case "on":
			on = true
		case "off":
			on = false
		default:
			return fmt.Errorf("usage: xpier directory-listing [on|off]")
		}
	}
	confPath := filepath.Join(nginx.NginxHome(), "nginx.conf")
	data, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}
	content := string(data)
	line := "\tautoindex on;\n"
	if on {
		if !strings.Contains(content, "autoindex") {
			content = strings.Replace(content, "    client_max_body_size 100m;", "    client_max_body_size 100m;\n"+line, 1)
		}
	} else {
		content = strings.ReplaceAll(content, line, "")
	}
	if err := os.WriteFile(confPath, []byte(content), 0o644); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		return fmt.Errorf("reload failed: %w", err)
	}
	fmt.Printf("directory listing %s\n", map[bool]string{true: "ON", false: "OFF"}[on])
	return nil
}

// cmdForget removes the current directory from parked paths (and unlinks it).
func cmdForget(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	// Unlink any site whose path is this directory (key may differ from basename).
	for name, site := range sites.Sites {
		if p, err := filepath.Abs(site.Path); err == nil && p == absCwd {
			delete(sites.Sites, name)
			nginx.RemoveSiteNginxConfig(name)
			break
		}
	}
	filtered := sites.Parked[:0]
	for _, p := range sites.Parked {
		if p != absCwd {
			filtered = append(filtered, p)
		}
	}
	sites.Parked = filtered
	if err := sites.Save(); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	fmt.Printf("forgot %s\n", absCwd)
	return nil
}

// --- Node version isolation (via nvm) ---

func cmdIsolateNode(args []string) error {
	fs := flag.NewFlagSet("isolate-node", flag.ExitOnError)
	siteFlag := fs.String("site", "", "site name (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier isolate-node <version> [--site x]")
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name, site, err := sitepkg.ResolveSite(sites, *siteFlag)
	if err != nil {
		return err
	}
	site.Node = fs.Arg(0)
	sites.Sites[name] = site
	if err := sites.Save(); err != nil {
		return err
	}
	fmt.Printf("%s node -> %s\n", store.SiteDomain(sites, name), site.Node)
	return nil
}

func cmdUnisolateNode(args []string) error {
	fs := flag.NewFlagSet("unisolate-node", flag.ExitOnError)
	siteFlag := fs.String("site", "", "site name (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	name, site, err := sitepkg.ResolveSite(sites, *siteFlag)
	if err != nil {
		return err
	}
	site.Node = ""
	sites.Sites[name] = site
	if err := sites.Save(); err != nil {
		return err
	}
	fmt.Printf("%s node unisolated\n", store.SiteDomain(sites, name))
	return nil
}

func cmdIsolatedNode(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	names := make([]string, 0)
	for name, site := range sites.Sites {
		if site.Node != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Println("no node-isolated sites")
		return nil
	}
	for _, name := range names {
		fmt.Printf("  %-30s node %s\n", store.SiteDomain(sites, name), sites.Sites[name].Node)
	}
	return nil
}

func cmdNode(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("usage: xpier node [--site x] args    run with the site's Node")
		return nil
	}
	siteName, passthrough := sitepkg.ExtractSiteFlag(args)
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := sitepkg.ResolveSite(sites, siteName)
	if err != nil {
		return err
	}
	ver := site.Node
	if ver == "" {
		return fmt.Errorf("site has no node pin; run `xpier isolate-node <version>`")
	}
	// nvm is a shell function; run through a login shell. Quote every
	// passthrough arg so spaces/quotes are not re-parsed by the shell.
	quoted := make([]string, 0, len(passthrough))
	for _, a := range passthrough {
		quoted = append(quoted, strconv.Quote(a))
	}
	script := fmt.Sprintf("source \"$NVM_DIR/nvm.sh\" 2>/dev/null || source \"$HOME/.nvm/nvm.sh\" 2>/dev/null; nvm exec %s node %s", ver, strings.Join(quoted, " "))
	cmd := exec.Command("bash", "-lc", script)
	cmd.Dir = site.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("node@%s exited: %w", ver, err)
	}
	return nil
}

// cmdIni opens a PHP version's php.ini (Herd's `herd ini` equivalent).
func cmdIni(args []string) error {
	fs := flag.NewFlagSet("ini", flag.ExitOnError)
	ver := fs.String("php", nginx.DefaultPhpVersion(), "php version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ini := filepath.Join(store.BrewPrefix(), "etc", "php", *ver, "php.ini")
	if !store.FileExists(ini) {
		return fmt.Errorf("php.ini not found at %s", ini)
	}
	return service.ShowConfig(ini)
}

// cmdCompletion prints a basic shell completion for subcommands.
func cmdCompletion(args []string) error {
	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}
	cmds := []string{"init init:fresh laravel:update sync doctor status app app:init app:status up app:up down app:down start app:start restart app:restart log app:log logs app:logs url app:url install refresh link unlink park forget paths sites sites:up sites:down site:php use php:list php:install php:update ext:install which which-php isolate unisolate isolated php composer debug coverage open edit site-information tld loopback links parked secure unsecure secured proxy proxies unproxy db:install db:start db:stop db:create db share shares share:stop fetch-share-url mail:up mail:down mail xdebug debug:start debug:stop tinker directory-listing isolate-node unisolate-node isolated-node node completion services services:stop services:start services:available services:versions services:create service ini"}
	switch shell {
	case "zsh":
		fmt.Printf("#compdef xpier\n_xpier() { compadd %s }\ncompdef _xpier xpier\n", cmds)
	default:
		fmt.Printf("_xpier() { local cur=${COMP_WORDS[COMP_CWORD]}; COMPREPLY=( $(compgen -W \"%s\" -- $cur) ); }\ncomplete -F _xpier xpier\n", cmds)
	}
	return nil
}

// cmdLaravelUpdate upgrades laravel/framework in the current site via
// composer (Herd's laravel:update).
func cmdLaravelUpdate(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	_, site, err := sitepkg.ResolveSite(sites, "")
	if err != nil {
		return err
	}
	if !store.FileExists(filepath.Join(site.Path, "composer.json")) {
		return fmt.Errorf("no composer.json in %s", site.Path)
	}
	cmd := exec.Command("composer", "update", "laravel/framework", "--with-all-dependencies")
	cmd.Dir = site.Path
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("composer update: %w", err)
	}
	return nil
}

// cmdFetchShareURL prints the public URL of a running share (the state
// file written by the background tunnel is the source of truth).
func cmdFetchShareURL(args []string) error {
	site := "default"
	if len(args) > 0 {
		site = args[0]
	}
	st, err := share.LoadShareState(site)
	if err != nil {
		return fmt.Errorf("no share for %q running (start with `xpier share %s`)", site, site)
	}
	if !store.ProcAlive(st.PID, "--url "+st.Target) {
		return fmt.Errorf("share %q is stale (pid %d no longer ours); run `xpier share %s` again", site, st.PID, site)
	}
	fmt.Println(st.URL)
	return nil
}
