package share

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

type ShareState struct {
	Site   string `json:"site"`
	PID    int    `json:"pid"`
	URL    string `json:"url"`
	Target string `json:"target"`
	Log    string `json:"log"`
	// Kind is the tunnel backend: "cloudflared" or "localhost-run" (ssh).
	Kind string `json:"kind,omitempty"`
}

// aliveMarker is the argv fragment identifying a live tunnel process.
func aliveMarker(st *ShareState) string {
	if st.Kind == "localhost-run" {
		return "nokey@localhost.run"
	}
	return "--url " + st.Target
}

func ShareStatePath(site string) string {
	return filepath.Join(store.XpierHome(), "servers", "share-"+site+".json")
}

func LoadShareState(site string) (*ShareState, error) {
	data, err := os.ReadFile(ShareStatePath(site))
	if err != nil {
		return nil, err
	}
	var st ShareState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func SaveShareState(st *ShareState) error {
	if err := os.MkdirAll(filepath.Dir(ShareStatePath(st.Site)), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ShareStatePath(st.Site), data, 0o644)
}

func CloudflaredBin() string {
	if p := filepath.Join(store.BrewPrefix(), "bin", "cloudflared"); store.FileExists(p) {
		return p
	}
	return "/usr/local/bin/cloudflared"
}

var trycloudflareRe = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)
var lhostRunRe = regexp.MustCompile(`https://[a-z0-9-]+\.(lhost\.run|lhr\.life)`)

// sshForwardSpec builds the `ssh -R` spec for localhost.run:
// <subdomain>:<port>:localhost:<hostport> (empty subdomain = random name).
func sshForwardSpec(target, subdomain string) string {
	hostPort := strings.TrimPrefix(strings.TrimPrefix(target, "http://"), "https://")
	port := hostPort
	if i := strings.LastIndex(hostPort, ":"); i >= 0 {
		port = hostPort[i+1:]
	}
	protoPort := "80"
	if strings.HasPrefix(target, "https://") {
		protoPort = "443"
	}
	return subdomain + ":" + protoPort + ":localhost:" + port
}

// startSSHTunnel shares via localhost.run (ssh -R): no account needed, and
// --domain picks a stable subdomain like https://myapp.lhost.run.
func startSSHTunnel(key, target, subdomain string) (string, error) {
	if _, err := exec.LookPath("ssh"); err != nil {
		return "", fmt.Errorf("ssh not found (localhost-run backend needs ssh)")
	}
	forward := sshForwardSpec(target, subdomain)
	logPath := filepath.Join(store.XpierHome(), "logs", "share-"+key+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	defer logFile.Close()
	cmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=accept-new", "-o", "ServerAliveInterval=30",
		"-R", forward, "nokey@localhost.run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start ssh tunnel: %w", err)
	}
	deadline := time.Now().Add(25 * time.Second)
	tunnelURL := ""
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(logPath)
		if m := lhostRunRe.FindString(string(data)); m != "" {
			tunnelURL = m
			break
		}
		if !store.PidAlive(cmd.Process.Pid) {
			break
		}
		time.Sleep(400 * time.Millisecond)
	}
	if tunnelURL == "" {
		store.KillGroup(cmd.Process.Pid, syscall.SIGKILL)
		return "", fmt.Errorf("localhost.run did not produce a URL; see %s", logPath)
	}
	st := &ShareState{Site: key, PID: cmd.Process.Pid, URL: tunnelURL, Target: target, Log: logPath, Kind: "localhost-run"}
	if err := SaveShareState(st); err != nil {
		return "", err
	}
	return tunnelURL, nil
}

// waitTunnelRegistered waits until cloudflared's log shows it connected to
// the Cloudflare edge (local and reliable signal).
func waitTunnelRegistered(logPath string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(logPath)
		if strings.Contains(string(data), "Registered tunnel connection") {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// verifyPublicURL curls the public URL (IPv4; this machine's IPv6 path to the
// Cloudflare edge is unreliable). Returns the HTTP code or "000".
func verifyPublicURL(url string) string {
	for i := 0; i < 2; i++ {
		code, err := store.RunOut("curl", "-4", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "8", url)
		if err == nil && code != "" && code != "000" {
			return code
		}
		time.Sleep(2 * time.Second)
	}
	return "000"
}

// startTunnel starts a managed cloudflared tunnel to a target URL and returns
// the public URL once it is live. insecure disables origin TLS verification
// (for self-signed dev certs like vite's basic-ssl).
func startTunnel(key, target string, insecure bool) (string, error) {
	bin := CloudflaredBin()
	if err := store.EnsureBrewPackage(bin, "cloudflared", "cloudflared"); err != nil {
		return "", err
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "share-"+key+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	defer logFile.Close()
	args := []string{"tunnel", "--url", target}
	if insecure {
		args = append(args, "--no-tls-verify")
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start cloudflared: %w", err)
	}
	tunnelURL := ""
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		data, _ := os.ReadFile(logPath)
		if m := trycloudflareRe.FindString(string(data)); m != "" {
			tunnelURL = m
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if tunnelURL == "" {
		store.KillGroup(cmd.Process.Pid, syscall.SIGKILL)
		return "", fmt.Errorf("cloudflared did not produce a tunnel URL; see %s", logPath)
	}
	if !waitTunnelRegistered(logPath, 20*time.Second) {
		store.KillGroup(cmd.Process.Pid, syscall.SIGKILL)
		return "", fmt.Errorf("cloudflared did not register; see %s", logPath)
	}
	st := &ShareState{Site: key, PID: cmd.Process.Pid, URL: tunnelURL, Target: target, Log: logPath}
	if err := SaveShareState(st); err != nil {
		return "", err
	}
	return tunnelURL, nil
}

// stopShareByKey stops a managed tunnel by its state key.
func stopShareByKey(key string) {
	st, err := LoadShareState(key)
	if err != nil {
		return
	}
	if store.ProcAlive(st.PID, aliveMarker(st)) {
		store.KillGroup(st.PID, syscall.SIGTERM)
		time.Sleep(300 * time.Millisecond)
		if store.PidAlive(st.PID) {
			store.KillGroup(st.PID, syscall.SIGKILL)
		}
	}
	os.Remove(ShareStatePath(key))
}

// probeURL returns the HTTP status code for a URL, or "" on failure.
func probeURL(u string) string {
	out, err := exec.Command("curl", "-sk", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "3", u).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// detectOriginProto probes whether a local port serves HTTP or HTTPS
// (e.g. vite with basic-ssl is HTTPS-only).
func detectOriginProto(port string) string {
	if code := probeURL("http://127.0.0.1:" + port); code != "" && code != "000" {
		return "http"
	}
	if code := probeURL("https://127.0.0.1:" + port); code != "" && code != "000" {
		return "https"
	}
	return "http"
}

func CmdShare(args []string) error {
	fs := flag.NewFlagSet("share", flag.ExitOnError)
	backend := fs.String("backend", "cloudflared", "tunnel backend (cloudflared | localhost-run)")
	port := fs.String("port", "", "share an existing local port (no site needed)")
	https := fs.Bool("https", false, "origin uses HTTPS (e.g. vite basic-ssl dev server)")
	domain := fs.String("domain", "", "stable subdomain for localhost-run backend (e.g. myapp -> https://myapp.lhost.run)")
	localhost := fs.Bool("localhost", false, "shorthand for --backend localhost-run (no account needed)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *localhost {
		*backend = "localhost-run"
	}
	siteName := fs.Arg(0)
	key := siteName
	url := "http://127.0.0.1:80"
	insecure := false
	s, err := store.LoadSites()
	if err != nil {
		return err
	}
	if *port != "" {
		// Share a running local server on an arbitrary port.
		key = "port-" + *port
		proto := "http"
		if *https {
			proto = "https"
		} else {
			proto = detectOriginProto(*port)
			if proto == "https" {
				fmt.Printf("origin on port %s is HTTPS, sharing with https\n", *port)
			}
		}
		url = proto + "://127.0.0.1:" + *port
		insecure = proto == "https"
		siteName = ""
	} else if siteName != "" {
		site, ok := s.Sites[siteName]
		if !ok {
			return fmt.Errorf("site %s is not linked", siteName)
		}
		if site.Driver == "hyperf" {
			url = "http://127.0.0.1:" + nginx.HyperfPort(site)
		} else {
			url = "http://" + store.SiteDomain(s, siteName) + "/"
		}
	}
	if key == "" {
		key = "default"
	}
	if st, err := LoadShareState(key); err == nil && store.ProcAlive(st.PID, aliveMarker(st)) {
		fmt.Printf("already sharing: %s (pid %d)\n", st.URL, st.PID)
		return nil
	}
	switch *backend {
	case "localhost-run":
		tunnelURL, err := startSSHTunnel(key, url, *domain)
		if err != nil {
			return err
		}
		fmt.Printf("sharing %s -> %s\n", url, tunnelURL)
		fmt.Println("status: `xpier share:list` | stop: `xpier share:stop " + key + "`")
		return nil
	case "cloudflared":
		tunnelURL, err := startTunnel(key, url, insecure)
		if err != nil {
			return err
		}
		// Map the tunnel host to the site so nginx serves it despite the
		// public Host header cloudflared forwards.
		if siteName != "" {
			host := strings.TrimPrefix(strings.TrimPrefix(tunnelURL, "https://"), "http://")
			if i := strings.IndexByte(host, '/'); i >= 0 {
				host = host[:i]
			}
			if err := nginx.WriteSiteNginxConfigWithNames(s, siteName, []string{host}); err != nil {
				return err
			}
			if err := nginx.NginxReload(); err != nil {
				fmt.Printf("[warn] nginx reload failed: %v\n", err)
			}
		}
		fmt.Printf("sharing %s -> %s\n", url, tunnelURL)
		if code := verifyPublicURL(tunnelURL); code != "000" {
			fmt.Printf("live: %s (HTTP %s)\n", tunnelURL, code)
		} else {
			fmt.Printf("live (registered): %s\n", tunnelURL)
		}
		fmt.Println("status: `xpier shares` | stop: `xpier share:stop " + key + "`")
		return nil
	default:
		return fmt.Errorf("unknown backend %q (cloudflared | localhost-run)", *backend)
	}
}

func CmdShares(args []string) error {
	entries, err := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	if err != nil {
		fmt.Println(store.Paint("no shares"))
		return nil
	}
	found := false
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "share-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		site := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "share-"), ".json")
		st, err := LoadShareState(site)
		if err != nil {
			continue
		}
		found = true
		state := store.Paint("down")
		if store.PidAlive(st.PID) {
			state = store.Paint("up")
		}
		fmt.Printf("  %-12s %-6s %s (pid %d, %s)\n", site, state, st.URL, st.PID, st.Target)
	}
	if !found {
		fmt.Println(store.Paint("no shares"))
	}
	return nil
}

func CmdShareStop(args []string) error {
	keys := []string{}
	if len(args) > 0 {
		keys = append(keys, args[0])
	} else {
		entries, _ := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "share-") && strings.HasSuffix(e.Name(), ".json") {
				keys = append(keys, strings.TrimSuffix(strings.TrimPrefix(e.Name(), "share-"), ".json"))
			}
		}
	}
	if len(keys) == 0 {
		return fmt.Errorf("no shares to stop")
	}
	for _, site := range keys {
		stopShareByKey(site)
		// Remove the tunnel host from the site's nginx config.
		if s, err := store.LoadSites(); err == nil {
			if _, ok := s.Sites[site]; ok {
				nginx.WriteSiteNginxConfig(s, site)
				nginx.NginxReload()
			}
		}
		fmt.Printf("stopped share %s\n", site)
	}
	return nil
}
