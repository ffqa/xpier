package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

func CurrentUser() (*user.User, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		return user.Lookup(sudoUser)
	}
	return user.Current()
}

// store.UDPBusy reports whether something is listening on a UDP port (dnsmasq).
// EnsureWildcardCert generates a self-signed *.test wildcard cert if missing.
func EnsureWildcardCert(tld string) error {
	cert, key := nginx.CertPaths(tld)
	if store.FileExists(cert) && store.FileExists(key) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cert), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-keyout", key, "-out", cert, "-days", "3650",
		"-subj", "/CN=*."+tld,
		"-addext", "subjectAltName=DNS:*."+tld+",DNS:"+tld+",DNS:localhost")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("openssl: %v: %s", err, out)
	}
	return nil
}

func LaunchdDir() string { return "/Library/LaunchDaemons" }

func LaunchdPlist(label string, args ...string) string {
	outLog := filepath.Join(store.XpierHome(), "logs", label+".out.log")
	errLog := filepath.Join(store.XpierHome(), "logs", label+".err.log")
	argv := ""
	for _, a := range args {
		argv += fmt.Sprintf("    <string>%s</string>\n", a)
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
%s  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>ProcessType</key>
  <string>Interactive</string>
  <key>StandardOutPath</key>
  <string>%s</string>
  <key>StandardErrorPath</key>
  <string>%s</string>
</dict>
</plist>
`, label, argv, outLog, errLog)
}

func LaunchdPlistNginx() string {
	return LaunchdPlist("com.xpier.nginx", nginx.NginxBin(), "-c", filepath.Join(nginx.NginxHome(), "nginx.conf"), "-g", "daemon off;")
}

func LaunchdPlistDnsmasq() string {
	return LaunchdPlist("com.xpier.dnsmasq", DnsmasqBin(), "-C", store.DnsmasqConfPath(), "--keep-in-foreground")
}

func DnsmasqBin() string {
	// brew installs dnsmasq under sbin, not bin.
	for _, p := range []string{
		filepath.Join(store.BrewPrefix(), "opt", "dnsmasq", "sbin", "dnsmasq"),
		filepath.Join(store.BrewPrefix(), "sbin", "dnsmasq"),
		"/usr/local/opt/dnsmasq/sbin/dnsmasq",
	} {
		if store.FileExists(p) {
			return p
		}
	}
	return "/usr/local/sbin/dnsmasq"
}

// EnsureNginxSudoers writes a passwordless sudoers entry so the user can
// reload the root-owned nginx master without typing a password.
func EnsureNginxSudoers() error {
	u, err := CurrentUser()
	if err != nil {
		return err
	}
	confPath := filepath.Join(nginx.NginxHome(), "nginx.conf")
	content := fmt.Sprintf("%s ALL=(root) NOPASSWD: %s -s reload -c %s\n%s ALL=(root) NOPASSWD: %s -t -c %s\n",
		u.Username, nginx.NginxBin(), confPath, u.Username, nginx.NginxBin(), confPath)
	return os.WriteFile("/etc/sudoers.d/xpier", []byte(content), 0o440)
}

func LaunchctlBootstrap(label, plistPath string) error {
	// A job left over from an earlier failed install can crash-loop; boot it
	// out first so bootstrap starts from a clean slate.
	store.RunOutErr("launchctl", "bootout", "system/"+label)
	cmd := exec.Command("launchctl", "bootstrap", "system", plistPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := string(out)
	if strings.Contains(msg, "Bootstrap failed: 5") || strings.Contains(msg, "already bootstrapped") {
		return store.RunOutErr("launchctl", "kickstart", "-k", "system/"+label)
	}
	return fmt.Errorf("launchctl bootstrap %s: %v: %s", label, err, msg)
}
