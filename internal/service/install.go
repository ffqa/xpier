package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

// BrewAsUser runs brew as the real user: Homebrew refuses to run as root.
func BrewAsUser(args ...string) (string, error) {
	u := os.Getenv("SUDO_USER")
	if u == "" {
		u = os.Getenv("USER")
	}
	cmd := exec.Command("sudo", append([]string{"-u", u, "-H", "brew"}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func CmdInstall(args []string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("xpier install must run as root: use `sudo xpier install`")
	}
	for _, pkg := range []string{"nginx", "dnsmasq"} {
		if out, err := BrewAsUser("list", "--versions", pkg); err == nil && strings.Contains(out, pkg) {
			fmt.Printf("%s already installed\n", pkg)
			continue
		}
		fmt.Printf("installing %s via brew (this may take a minute)...\n", pkg)
		if out, err := BrewAsUser("install", pkg); err != nil {
			return fmt.Errorf("brew install %s: %v: %s", pkg, err, out)
		}
	}
	if err := checkPortConflicts(); err != nil {
		return err
	}
	fmt.Println("port checks passed")
	for _, d := range []string{"nginx", "nginx/conf.d", "dnsmasq", "fpm", "run", "certs", "logs"} {
		if err := os.MkdirAll(filepath.Join(store.XpierHome(), d), 0o755); err != nil {
			return err
		}
	}
	fmt.Println("writing nginx + dnsmasq configs...")
	if err := nginx.WriteNginxMainConfig(); err != nil {
		return err
	}
	if err := nginx.WriteDefaultSiteConfig(); err != nil {
		return err
	}
	// Regenerate site configs so they reference current paths.
	if sites, err := store.LoadSites(); err == nil {
		nginx.WriteAllSiteConfigs(sites)
	}
	tld := nginx.CurrentTLD()
	if err := store.WriteDnsmasqConfig(tld); err != nil {
		return err
	}
	if err := EnsureWildcardCert(tld); err != nil {
		return err
	}
	if err := EnsureNginxSudoers(); err != nil {
		return fmt.Errorf("write sudoers: %w", err)
	}
	if err := ChownHerdyHomeToUser(); err != nil {
		return err
	}
	fmt.Println("configs written, installing launchd daemons...")
	nginxPlist := filepath.Join(LaunchdDir(), "com.xpier.nginx.plist")
	dnsPlist := filepath.Join(LaunchdDir(), "com.xpier.dnsmasq.plist")
	if err := os.WriteFile(nginxPlist, []byte(LaunchdPlistNginx()), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(dnsPlist, []byte(LaunchdPlistDnsmasq()), 0o644); err != nil {
		return err
	}
	// Boot out any daemons from earlier names (com.herdy.*, com.pier.*) so
	// they do not keep holding ports while the new com.xpier.* daemons start.
	store.RunOutErr("launchctl", "bootout", "system/com.herdy.nginx")
	store.RunOutErr("launchctl", "bootout", "system/com.herdy.dnsmasq")
	store.RunOutErr("launchctl", "bootout", "system/com.pier.nginx")
	store.RunOutErr("launchctl", "bootout", "system/com.pier.dnsmasq")
	if err := LaunchctlBootstrap("com.xpier.nginx", nginxPlist); err != nil {
		return err
	}
	if err := LaunchctlBootstrap("com.xpier.dnsmasq", dnsPlist); err != nil {
		return err
	}
	fmt.Println("xpier install complete.")
	fmt.Println("next: cd <project> && xpier link && xpier sites:up")
	return nil
}

func checkPortConflicts() error {
	checks := []struct{ port, proto, what string }{
		{"80", "tcp", "nginx (HTTP)"},
		{"443", "tcp", "nginx (HTTPS)"},
		{"53", "udp", "dnsmasq (DNS)"},
	}
	for _, c := range checks {
		var busy bool
		var holder string
		if c.proto == "udp" {
			b, _ := store.UDPBusy(c.port)
			busy = b
			if b {
				out, _ := store.RunOut("lsof", "-nP", "-iUDP:"+c.port)
				holder = firstNonHeaderLine(out)
			}
		} else {
			b, _ := store.PortBusy(c.port)
			busy = b
			if b {
				out, _ := store.RunOut("lsof", "-nP", "-iTCP:"+c.port, "-sTCP:LISTEN")
				holder = firstNonHeaderLine(out)
			}
		}
		if busy {
			// A port held by our own daemon (e.g. a leftover com.xpier.nginx)
			// is not a conflict; bootstrap below will restart it cleanly.
			if isOurBinary(holder) {
				continue
			}
			return fmt.Errorf("port %s (%s) is already in use by:\n  %s\nThis is likely Herd. Stop it first (quit the Herd app or kill its nginx/dnsmasq), then re-run `sudo xpier install`.",
				c.port, c.what, holder)
		}
	}
	return nil
}

func isOurBinary(holder string) bool {
	// lsof -nP shows only the command name; resolve the pid to its full path.
	fields := strings.Fields(holder)
	if len(fields) < 2 {
		return false
	}
	cmdline, _ := store.RunOut("ps", "-o", "command=", "-p", fields[1])
	if cmdline == "" {
		return false
	}
	if strings.Contains(cmdline, nginx.NginxBin()) ||
		strings.Contains(cmdline, "/opt/nginx/bin/nginx") ||
		strings.Contains(cmdline, DnsmasqBin()) ||
		strings.Contains(cmdline, "/opt/dnsmasq/sbin/dnsmasq") {
		return true
	}
	// Herd ships nginx-x86/nginx-arm64; a plain nginx/dnsmasq is ours.
	herds := strings.Contains(cmdline, "nginx-x86") || strings.Contains(cmdline, "nginx-arm64")
	return !herds && (strings.Contains(cmdline, "nginx") || strings.Contains(cmdline, "dnsmasq"))
}

func firstNonHeaderLine(out string) string {
	for _, line := range strings.Split(out, "\n") {
		if line != "" && !strings.HasPrefix(line, "COMMAND") {
			return line
		}
	}
	return out
}

// ChownHerdyHomeToUser returns ownership of ~/.xpier files to the real user:
// configs written during `sudo xpier install` would otherwise be root-owned
// and uneditable by the user (nginx.conf, dnsmasq.conf, ...).
func ChownHerdyHomeToUser() error {
	u, err := CurrentUser()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	return filepath.Walk(store.XpierHome(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		return os.Chown(path, uid, gid)
	})
}
