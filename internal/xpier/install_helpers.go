package xpier

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

func currentUser() (*user.User, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		return user.Lookup(sudoUser)
	}
	return user.Current()
}

// udpBusy reports whether something is listening on a UDP port (dnsmasq).
func udpBusy(port string) (bool, error) {
	out, err := runOut("lsof", "-ti", "udp:"+port)
	if err != nil && out == "" {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

func dnsmasqConfPath() string {
	return filepath.Join(xpierHome(), "dnsmasq", "dnsmasq.conf")
}

func writeDnsmasqConfig(tld string) error {
	conf := fmt.Sprintf(`port=53
listen-address=127.0.0.1
bind-interfaces
no-resolv
address=/.%s/127.0.0.1
`, tld)
	if err := os.MkdirAll(filepath.Dir(dnsmasqConfPath()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dnsmasqConfPath(), []byte(conf), 0o644)
}

func certPaths(tld string) (string, string) {
	return filepath.Join(xpierHome(), "certs", "wildcard."+tld+".pem"),
		filepath.Join(xpierHome(), "certs", "wildcard."+tld+"-key.pem")
}

// ensureWildcardCert generates a self-signed *.test wildcard cert if missing.
func ensureWildcardCert(tld string) error {
	cert, key := certPaths(tld)
	if fileExists(cert) && fileExists(key) {
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

func launchdDir() string { return "/Library/LaunchDaemons" }

func launchdPlist(label string, args ...string) string {
	outLog := filepath.Join(xpierHome(), "logs", label+".out.log")
	errLog := filepath.Join(xpierHome(), "logs", label+".err.log")
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

func launchdPlistNginx() string {
	return launchdPlist("com.xpier.nginx", nginxBin(), "-c", filepath.Join(nginxHome(), "nginx.conf"), "-g", "daemon off;")
}

func launchdPlistDnsmasq() string {
	return launchdPlist("com.xpier.dnsmasq", dnsmasqBin(), "-C", dnsmasqConfPath(), "--keep-in-foreground")
}

func dnsmasqBin() string {
	// brew installs dnsmasq under sbin, not bin.
	for _, p := range []string{
		filepath.Join(brewPrefix(), "opt", "dnsmasq", "sbin", "dnsmasq"),
		filepath.Join(brewPrefix(), "sbin", "dnsmasq"),
		"/usr/local/opt/dnsmasq/sbin/dnsmasq",
	} {
		if fileExists(p) {
			return p
		}
	}
	return "/usr/local/sbin/dnsmasq"
}

// ensureNginxSudoers writes a passwordless sudoers entry so the user can
// reload the root-owned nginx master without typing a password.
func ensureNginxSudoers() error {
	u, err := currentUser()
	if err != nil {
		return err
	}
	confPath := filepath.Join(nginxHome(), "nginx.conf")
	content := fmt.Sprintf("%s ALL=(root) NOPASSWD: %s -s reload -c %s\n%s ALL=(root) NOPASSWD: %s -t -c %s\n",
		u.Username, nginxBin(), confPath, u.Username, nginxBin(), confPath)
	return os.WriteFile("/etc/sudoers.d/xpier", []byte(content), 0o440)
}

func launchctlBootstrap(label, plistPath string) error {
	// A job left over from an earlier failed install can crash-loop; boot it
	// out first so bootstrap starts from a clean slate.
	runOutErr("launchctl", "bootout", "system/"+label)
	cmd := exec.Command("launchctl", "bootstrap", "system", plistPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	msg := string(out)
	if strings.Contains(msg, "Bootstrap failed: 5") || strings.Contains(msg, "already bootstrapped") {
		return runOutErr("launchctl", "kickstart", "-k", "system/"+label)
	}
	return fmt.Errorf("launchctl bootstrap %s: %v: %s", label, err, msg)
}

func runOutErr(name string, args ...string) error {
	_, err := runOut(name, args...)
	return err
}
