package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"xpier/internal/store"
)

func NginxHome() string {
	return filepath.Join(store.XpierHome(), "nginx")
}

func NginxConfDir() string {
	return filepath.Join(NginxHome(), "conf.d")
}

func NginxBin() string {
	if p := filepath.Join(store.BrewPrefix(), "bin", "nginx"); store.FileExists(p) {
		return p
	}
	return "/usr/local/bin/nginx"
}

func SiteConfPath(name string) string {
	return filepath.Join(NginxConfDir(), name+".conf")
}

func FastcgiParamsPath() string {
	return filepath.Join(store.BrewPrefix(), "etc", "nginx", "fastcgi_params")
}

func WriteNginxMainConfig() error {
	user, err := store.CurrentUser()
	if err != nil {
		return err
	}
	conf := fmt.Sprintf(`user %s;
worker_processes auto;
error_log %s/error.log;
pid %s/nginx.pid;
events {
    worker_connections 1024;
}
http {
    include %s;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 100m;
    server_names_hash_bucket_size 256;
    include %s/*.conf;
}
`, user.Username, NginxHome(), NginxHome(), filepath.Join(store.BrewPrefix(), "etc", "nginx", "mime.types"), NginxConfDir())
	if err := os.MkdirAll(NginxHome(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(NginxHome(), "nginx.conf"), []byte(conf), 0o644)
}

// DefaultPhpVersion returns the pinned default (`xpier use`), else the
// highest brew php@X installed, else 8.2.
func DefaultPhpVersion() string {
	if sites, err := store.LoadSites(); err == nil && sites.DefaultPHP != "" {
		if isVersionString(sites.DefaultPHP) {
			return sites.DefaultPHP
		}
	}
	best := ""
	entries, err := os.ReadDir(filepath.Join(store.BrewPrefix(), "opt"))
	if err != nil {
		return "8.2"
	}
	for _, e := range entries {
		v := strings.TrimPrefix(e.Name(), "php@")
		if strings.HasPrefix(e.Name(), "php@") && isVersionString(v) {
			if compareVersionStrings(v, best) > 0 {
				best = v
			}
		}
	}
	if best == "" {
		return "8.2"
	}
	return best
}

func isVersionString(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
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

func compareVersionStrings(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &x)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &y)
		}
		if x != y {
			return x - y
		}
	}
	return 0
}

func WriteSiteNginxConfig(sites *store.Sites, name string) error {
	return WriteSiteNginxConfigWithNames(sites, name, nil)
}

// WriteSiteNginxConfigWithNames writes a site's nginx config. extra names are
// additional server_name values (used by `xpier share` so the tunnel host
// resolves to the site).
func WriteSiteNginxConfigWithNames(sites *store.Sites, name string, extra []string) error {
	site, ok := sites.Sites[name]
	if !ok {
		return fmt.Errorf("site %s not linked", name)
	}
	php := site.PHP
	if php == "" {
		php = DefaultPhpVersion()
	}
	domain := store.SiteDomain(sites, name)
	root := store.SiteRoot(site)
	cert := filepath.Join(store.XpierHome(), "certs", "wildcard."+sites.TLD+".pem")
	certKey := filepath.Join(store.XpierHome(), "certs", "wildcard."+sites.TLD+"-key.pem")
	// Prefer a per-domain cert (signed via `sudo xpier secure <domain>`) when
	// one exists; the *.test wildcard does not cover multi-label hosts like
	// img.test28.test.
	if dc, dk := DomainCertPaths(domain); store.FileExists(dc) && store.FileExists(dk) {
		cert, certKey = dc, dk
	}

	serverNames := domain
	for _, n := range extra {
		if n != "" {
			serverNames += " " + n
		}
	}

	var conf strings.Builder
	conf.WriteString("server {\n")
	// Loopback-only: .test resolves to 127.0.0.1, and this keeps 80/443 free
	// on other interfaces for tools like Tailscale funnel.
	conf.WriteString("    listen 127.0.0.1:80;\n")
	httpsOn := site.Secure == nil || *site.Secure
	if httpsOn {
		conf.WriteString("    listen 127.0.0.1:443 ssl;\n")
	}
	fmt.Fprintf(&conf, "    server_name %s;\n", serverNames)
	if httpsOn {
		fmt.Fprintf(&conf, "    ssl_certificate     %s;\n", cert)
		fmt.Fprintf(&conf, "    ssl_certificate_key %s;\n", certKey)
	}
	fmt.Fprintf(&conf, "    root \"%s\";\n", root)
	conf.WriteString("    index index.php index.html;\n")
	if site.Driver == "hyperf" {
		port := HyperfPort(site)
		conf.WriteString("    location / {\n")
		fmt.Fprintf(&conf, "        proxy_pass http://127.0.0.1:%s;\n", port)
		conf.WriteString("        proxy_http_version 1.1;\n")
		conf.WriteString("        proxy_set_header Host $host;\n")
		conf.WriteString("        proxy_set_header X-Real-IP $remote_addr;\n")
		conf.WriteString("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n")
		conf.WriteString("        proxy_set_header Upgrade $http_upgrade;\n")
		conf.WriteString("        proxy_set_header Connection \"upgrade\";\n")
		conf.WriteString("    }\n")
	} else if site.Driver == "laravel" || site.Driver == "php" {
		sock := filepath.Join(store.XpierHome(), "run", fmt.Sprintf("php-fpm-%s.sock", php))
		conf.WriteString("    location / {\n")
		conf.WriteString("        try_files $uri $uri/ /index.php?$query_string;\n")
		conf.WriteString("    }\n")
		conf.WriteString("    location ~ \\.php$ {\n")
		fmt.Fprintf(&conf, "        fastcgi_pass unix:%s;\n", sock)
		fmt.Fprintf(&conf, "        include %s;\n", FastcgiParamsPath())
		conf.WriteString("        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n")
		conf.WriteString("    }\n")
	} else {
		conf.WriteString("    location / {\n")
		conf.WriteString("        try_files $uri $uri/ =404;\n")
		conf.WriteString("    }\n")
	}
	conf.WriteString("}\n")

	if err := os.MkdirAll(NginxConfDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(SiteConfPath(name), []byte(conf.String()), 0o644)
}

func RemoveSiteNginxConfig(name string) error {
	err := os.Remove(SiteConfPath(name))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// CurrentTLD returns the active TLD (defaults to "test"). Certificates and
// configs must be derived from it so `xpier tld dev` does not break TLS.
func CurrentTLD() string {
	tld := "test"
	if sites, err := store.LoadSites(); err == nil && sites.TLD != "" {
		tld = sites.TLD
	}
	return tld
}

// WriteDefaultSiteConfig returns 404 for any host that is not a registered
// site, so an unlinked domain can never fall through to another site.
func WriteDefaultSiteConfig() error {
	cert, certKey := CertPaths(CurrentTLD())
	conf := fmt.Sprintf(`server {
    listen 127.0.0.1:80 default_server;
    listen 127.0.0.1:443 ssl default_server;
    server_name _;
    ssl_certificate     %s;
    ssl_certificate_key %s;
    return 404;
}
`, cert, certKey)
	if err := os.MkdirAll(NginxConfDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(NginxConfDir(), "00-default.conf"), []byte(conf), 0o644)
}

// NginxReload reloads the launchd-managed nginx master via a passwordless
// sudoers entry installed by `xpier install`. The -c flag is required so
// nginx reads our pid file (defaults would target /usr/local/var/run).
func NginxReload() error {
	cmd := exec.Command("sudo", "-n", NginxBin(), "-s", "reload", "-c", filepath.Join(NginxHome(), "nginx.conf"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func LoadManifestFrom(dir string) (*store.Manifest, error) {
	manifestPath, _ := store.ResolvePaths(dir)
	return store.LoadManifest(manifestPath)
}

// ServerPorts best-effort reads config/autoload/server.php for name -> port pairs.
func ServerPorts(dir string) map[string]string {
	data, err := os.ReadFile(filepath.Join(dir, "config", "autoload", "server.php"))
	if err != nil {
		return nil
	}
	names := serverNameRe.FindAllStringSubmatch(string(data), -1)
	ports := serverPortRe.FindAllStringSubmatch(string(data), -1)
	if len(names) == 0 || len(ports) == 0 {
		return nil
	}
	out := make(map[string]string)
	for i := 0; i < len(names) && i < len(ports); i++ {
		out[names[i][1]] = ports[i][1]
	}
	return out
}

var (
	serverNameRe = regexp.MustCompile(`'name'\s*=>\s*'([a-z0-9_]+)'`)
	serverPortRe = regexp.MustCompile(`'port'\s*=>\s*(\d+)`)
)

// HyperfPort returns the proxy port for a hyperf site.
func HyperfPort(site store.Site) string {
	ports := ServerPorts(site.Path)
	if p, ok := ports["http"]; ok {
		return p
	}
	return "9501"
}

func CertPaths(tld string) (string, string) {
	return filepath.Join(store.XpierHome(), "certs", "wildcard."+tld+".pem"),
		filepath.Join(store.XpierHome(), "certs", "wildcard."+tld+"-key.pem")
}

func DomainCertPaths(domain string) (string, string) {
	return filepath.Join(store.XpierHome(), "certs", domain+".pem"),
		filepath.Join(store.XpierHome(), "certs", domain+"-key.pem")
}

// WriteAllSiteConfigs regenerates nginx configs for every registered site.
func WriteAllSiteConfigs(sites *store.Sites) error {
	if err := WriteDefaultSiteConfig(); err != nil {
		return err
	}
	for name := range sites.Sites {
		if err := WriteSiteNginxConfig(sites, name); err != nil {
			return err
		}
	}
	return nil
}
