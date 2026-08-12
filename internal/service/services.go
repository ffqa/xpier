package service

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/share"
	"xpier/internal/store"
)

func DnsmasqRunning() bool {
	// lsof sometimes fails to report dnsmasq's UDP 53 socket on macOS;
	// detect the process instead.
	out, err := store.RunOut("pgrep", "-f", DnsmasqBin())
	return err == nil && strings.TrimSpace(out) != ""
}

func XpierServiceStatus() {
	nginxUp := false
	if b, _ := store.PortBusy("80"); b {
		nginxUp = true
	}
	dnsUp := DnsmasqRunning()
	fmt.Printf("nginx:   %s\n", store.UpDown(nginxUp))
	fmt.Printf("dnsmasq: %s\n", store.UpDown(dnsUp))
	entries, _ := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	fpm := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			if FpmRunning(ver) {
				fpm = append(fpm, ver)
			}
		}
	}
	if len(fpm) == 0 {
		fmt.Println("php-fpm: " + store.Paint("down"))
	} else {
		fmt.Printf("php-fpm: %s (%s)\n", store.Paint("up"), strings.Join(fpm, ", "))
	}
	// Shares (cloudflared tunnels)
	entries, _ = os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	shares := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "share-") && strings.HasSuffix(e.Name(), ".json") {
			site := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "share-"), ".json")
			if st, err := share.LoadShareState(site); err == nil && store.PidAlive(st.PID) {
				shares = append(shares, fmt.Sprintf("%s(%s)", site, st.URL))
			}
		}
	}
	if len(shares) == 0 {
		fmt.Println("share:   " + store.Paint("none") + " (cloudflared)")
	} else {
		fmt.Printf("share:   %s: %s\n", store.Paint("up"), strings.Join(shares, ", "))
	}
}

func CmdServices(args []string) error {
	XpierServiceStatus()
	return nil
}

func StopDaemon(label string) error {
	out, err := store.RunOut("launchctl", "bootout", "system/"+label)
	if err != nil {
		return fmt.Errorf("bootout %s: %v: %s", label, err, out)
	}
	return nil
}

func StartDaemon(label string) error {
	return LaunchctlBootstrap(label, filepath.Join(LaunchdDir(), label+".plist"))
}

func RestartDaemon(label string) error {
	store.RunOutErr("launchctl", "bootout", "system/"+label)
	return StartDaemon(label)
}

func StopDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := StopDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

func StartDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := StartDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

// ShowConfig prints the path and opens the file in an editor when available,
// otherwise prints the content.
func ShowConfig(path string) error {
	fmt.Printf("config: %s\n", path)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if store.FileExists("/usr/local/bin/code") {
			editor = "/usr/local/bin/code"
		}
	}
	if editor != "" {
		return store.RunOutErr(editor, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

// CmdService manages a single service: nginx / dnsmasq / php-fpm[-<ver>].
func CmdService(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: xpier service <nginx|dnsmasq|php-fpm|php-fpm-8.2> <status|config|configtest|reload|start|stop|restart>")
	}
	name, action := args[0], args[1]
	ver := ""
	if strings.HasPrefix(name, "php-fpm") {
		ver = strings.TrimPrefix(name, "php-fpm-")
		if ver == name || ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
		name = "php-fpm"
	}
	switch name {
	case "nginx":
		switch action {
		case "status":
			b, _ := store.PortBusy("80")
			fmt.Printf("nginx: %s\n", store.UpDown(b))
		case "config":
			return ShowConfig(filepath.Join(nginx.NginxHome(), "nginx.conf"))
		case "configtest":
			out, err := store.RunOut("sudo", "-n", nginx.NginxBin(), "-t", "-c", filepath.Join(nginx.NginxHome(), "nginx.conf"))
			if err != nil {
				return fmt.Errorf("nginx configtest failed: %v: %s", err, out)
			}
			fmt.Println("nginx configuration test successful")
		case "reload":
			if err := nginx.NginxReload(); err != nil {
				return err
			}
			fmt.Println("nginx reloaded")
		case "start", "stop", "restart":
			if os.Geteuid() != 0 {
				return fmt.Errorf("nginx is a launchd daemon; run `sudo xpier service nginx %s`", action)
			}
			var err error
			switch action {
			case "start":
				err = StartDaemon("com.xpier.nginx")
			case "stop":
				err = StopDaemon("com.xpier.nginx")
			case "restart":
				err = RestartDaemon("com.xpier.nginx")
			}
			if err != nil {
				return err
			}
			fmt.Printf("nginx %s\n", action)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	case "dnsmasq":
		switch action {
		case "status":
			fmt.Printf("dnsmasq: %s\n", store.UpDown(DnsmasqRunning()))
		case "config":
			return ShowConfig(store.DnsmasqConfPath())
		case "configtest":
			out, err := store.RunOut(DnsmasqBin(), "--test", "-C", store.DnsmasqConfPath())
			if err != nil {
				return fmt.Errorf("dnsmasq configtest failed: %v: %s", err, out)
			}
			fmt.Println("dnsmasq configuration test successful")
		case "start", "stop", "restart":
			if os.Geteuid() != 0 {
				return fmt.Errorf("dnsmasq is a launchd daemon; run `sudo xpier service dnsmasq %s`", action)
			}
			var err error
			switch action {
			case "start":
				err = StartDaemon("com.xpier.dnsmasq")
			case "stop":
				err = StopDaemon("com.xpier.dnsmasq")
			case "restart":
				err = RestartDaemon("com.xpier.dnsmasq")
			}
			if err != nil {
				return err
			}
			fmt.Printf("dnsmasq %s\n", action)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	case "php-fpm":
		switch action {
		case "status":
			fmt.Printf("php-fpm %s: %s\n", ver, store.UpDown(FpmRunning(ver)))
		case "config":
			return ShowConfig(FpmConfPath(ver))
		case "start":
			return FpmUp(ver)
		case "stop":
			return FpmDown(ver)
		case "restart":
			FpmDown(ver)
			return FpmUp(ver)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	default:
		return fmt.Errorf("unknown service %q (nginx|dnsmasq|php-fpm)", name)
	}
	return nil
}

func CmdServicesStop(args []string) error {
	entries, _ := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			FpmDown(ver)
		}
	}
	if os.Geteuid() == 0 {
		return StopDaemons()
	}
	fmt.Println("php-fpm stopped; nginx/dnsmasq still running (run `sudo xpier services:stop` to stop them too)")
	return nil
}

func CmdServicesStart(args []string) error {
	if os.Geteuid() == 0 {
		if err := StartDaemons(); err != nil {
			return err
		}
		fmt.Println("nginx + dnsmasq daemons started")
	} else {
		fmt.Println("starting php-fpm for linked sites (run `sudo xpier services:start` to also start nginx/dnsmasq)")
	}
	return startSiteFpms()
}

func startSiteFpms() error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
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
		if err := FpmUp(ver); err != nil {
			fmt.Printf("[warn] %v\n", err)
		}
	}
	return nil
}

var knownServices = []string{"mysql", "mariadb", "redis", "postgresql@14", "postgresql@16", "mailpit", "cloudflared"}

func isHelpArg(args []string) bool {
	return len(args) > 0 && (args[0] == "-h" || args[0] == "--help")
}

// CmdServicesAvailable lists the services xpier can install via brew and
// whether each is already installed (Herd Pro's services:available).
func CmdServicesAvailable(args []string) error {
	fmt.Println("available services (brew):")
	for _, svc := range knownServices {
		status := "-"
		if out, err := store.RunOut("brew", "list", "--versions", svc); err == nil && out != "" {
			status = out
		}
		fmt.Printf("  %-16s %s\n", svc, status)
	}
	return nil
}

// CmdServicesVersions lists installed versions of known services.
func CmdServicesVersions(args []string) error {
	found := false
	for _, svc := range knownServices {
		if out, err := store.RunOut("brew", "list", "--versions", svc); err == nil && out != "" {
			fmt.Printf("  %-16s %s\n", svc, out)
			found = true
		}
	}
	if !found {
		fmt.Println("no known services installed")
	}
	return nil
}

// CmdServicesCreate installs a service via brew and starts it
// (lightweight stand-in for Herd Pro's managed database services).
func CmdServicesCreate(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier services:create <mysql|mariadb|redis|postgres|mailpit>")
		return nil
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: xpier services:create <mysql|mariadb|redis|postgres>")
	}
	aliases := map[string]string{
		"mysql": "mysql", "mariadb": "mariadb", "maria": "mariadb",
		"redis": "redis", "postgres": "postgresql@16", "postgresql": "postgresql@16",
		"postgresql@16": "postgresql@16", "postgresql@14": "postgresql@14",
		"mailpit": "mailpit",
	}
	formula, ok := aliases[args[0]]
	if !ok {
		return fmt.Errorf("unknown service %q (mysql|mariadb|redis|postgres|mailpit)", args[0])
	}
	if out, err := BrewAsUser("list", "--versions", formula); err == nil && strings.Contains(out, formula) {
		fmt.Printf("%s already installed\n", formula)
	} else {
		fmt.Printf("installing %s via brew (progress below)...\n", formula)
		if err := store.RunOutLive("brew", "install", formula); err != nil {
			return fmt.Errorf("brew install %s failed: %w (run it manually to see the full log)", formula, err)
		}
	}
	if err := store.RunOutLive("brew", "services", "start", formula); err != nil {
		return fmt.Errorf("brew services start %s: %w", formula, err)
	}
	fmt.Printf("%s installed and started\n", formula)
	return nil
}

var safeExtNameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// CmdExtInstall installs a PHP extension for a PHP version via brew
// (shivammathur/extensions tap, e.g. `xpier ext:install swoole --php 8.4`).
func CmdExtInstall(args []string) error {
	fs := flag.NewFlagSet("ext:install", flag.ExitOnError)
	ver := fs.String("php", nginx.DefaultPhpVersion(), "php version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier ext:install <swoole|redis|xdebug|...> [--php 8.4]")
	}
	ext := fs.Arg(0)
	if !safeExtNameRe.MatchString(ext) {
		return fmt.Errorf("invalid extension name %q", ext)
	}
	// Tap may already exist; failure is not fatal (brew install would tap it too).
	if err := store.RunOutLive("brew", "tap", "shivammathur/extensions"); err != nil {
		fmt.Println("[warn] tap shivammathur/extensions: " + err.Error())
	}
	fmt.Printf("installing %s@%s via brew (progress below)...\n", ext, *ver)
	if err := store.RunOutLive("brew", "install", "shivammathur/extensions/"+ext+"@"+*ver); err != nil {
		return fmt.Errorf("brew install %s@%s failed: %w (run it manually to see the full log)", ext, *ver, err)
	}
	fmt.Printf("%s@%s installed (restart php-fpm to load it: `xpier service php-fpm-%s restart`)\n", ext, *ver, *ver)
	return nil
}

// CmdPhpInstall installs a PHP version via brew (shivammathur tap).
func CmdPhpInstall(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier php:install <8.2|8.3|...>    install a PHP version via brew")
		return nil
	}
	if len(args) < 1 {
		return fmt.Errorf("usage: xpier php:install <8.2|8.3|...>")
	}
	ver := args[0]
	if !store.SafePhpRe.MatchString(ver) {
		return fmt.Errorf("invalid php version %q", ver)
	}
	fmt.Printf("installing php@%s via brew (progress below)...\n", ver)
	if err := store.RunOutLive("brew", "install", "shivammathur/php/php@"+ver); err != nil {
		return fmt.Errorf("brew install php@%s failed: %w (run it manually to see the full log)", ver, err)
	}
	fmt.Printf("php@%s installed\n", ver)
	return nil
}

// CmdPhpUpdate upgrades an installed PHP version via brew (default: the
// global default version).
func CmdPhpUpdate(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier php:update [ver]    upgrade an installed PHP version")
		return nil
	}
	ver := nginx.DefaultPhpVersion()
	if len(args) > 0 {
		ver = args[0]
	}
	if !store.SafePhpRe.MatchString(ver) {
		return fmt.Errorf("invalid php version %q", ver)
	}
	fmt.Printf("upgrading php@%s via brew (progress below)...\n", ver)
	if err := store.RunOutLive("brew", "upgrade", "shivammathur/php/php@"+ver); err != nil {
		return fmt.Errorf("brew upgrade php@%s failed: %w (run it manually to see the full log)", ver, err)
	}
	fmt.Printf("php@%s upgraded\n", ver)
	return nil
}

// CmdPhpList lists installed PHP versions (brew opt/php@*), marking the
// global default (`xpier use`) and the auto-selected default.
func CmdPhpList(args []string) error {
	entries, err := os.ReadDir(filepath.Join(store.BrewPrefix(), "opt"))
	if err != nil {
		return fmt.Errorf("no PHP versions found under %s/opt", store.BrewPrefix())
	}
	type phpVer struct {
		ver  string
		bin  string
		full string
	}
	var list []phpVer
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "php@") {
			continue
		}
		ver := strings.TrimPrefix(name, "php@")
		if !isPhpVer(ver) {
			continue
		}
		bin := filepath.Join(store.BrewPrefix(), "opt", name, "bin", "php")
		full := ""
		if out, err := store.RunOut(bin, "-v"); err == nil {
			if first := strings.SplitN(out, "\n", 2)[0]; first != "" {
				full = first
			}
		}
		list = append(list, phpVer{ver, bin, full})
	}
	if len(list) == 0 {
		return fmt.Errorf("no PHP versions installed (brew install shivammathur/php/php@8.2)")
	}
	defaultVer := ""
	if sites, err := store.LoadSites(); err == nil {
		defaultVer = sites.DefaultPHP
	}
	if defaultVer == "" {
		defaultVer = nginx.DefaultPhpVersion()
	}
	sort.Slice(list, func(i, j int) bool {
		return comparePhpVer(list[i].ver, list[j].ver) < 0
	})
	fmt.Printf("%-8s %-5s %s\n", "VERSION", "DEFAULT", "PHP BINARY")
	for _, p := range list {
		mark := ""
		if p.ver == defaultVer {
			mark = "default"
		}
		fmt.Printf("%-8s %-5s %s\n", p.ver, mark, p.bin)
		if p.full != "" {
			fmt.Printf("         %s\n", p.full)
		}
	}
	return nil
}

func isPhpVer(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func comparePhpVer(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	var x, y int
	fmt.Sscanf(pa[0], "%d", &x)
	fmt.Sscanf(pb[0], "%d", &y)
	if x != y {
		return x - y
	}
	fmt.Sscanf(pa[1], "%d", &x)
	fmt.Sscanf(pb[1], "%d", &y)
	return x - y
}
