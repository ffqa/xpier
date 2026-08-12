package xpier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func nginxHome() string {
	return filepath.Join(xpierHome(), "nginx")
}

func nginxConfDir() string {
	return filepath.Join(nginxHome(), "conf.d")
}

func nginxBin() string {
	if p := filepath.Join(brewPrefix(), "bin", "nginx"); fileExists(p) {
		return p
	}
	return "/usr/local/bin/nginx"
}

func siteConfPath(name string) string {
	return filepath.Join(nginxConfDir(), name+".conf")
}

func fastcgiParamsPath() string {
	return filepath.Join(brewPrefix(), "etc", "nginx", "fastcgi_params")
}

func writeNginxMainConfig() error {
	user, err := currentUser()
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
`, user.Username, nginxHome(), nginxHome(), filepath.Join(brewPrefix(), "etc", "nginx", "mime.types"), nginxConfDir())
	return os.WriteFile(filepath.Join(nginxHome(), "nginx.conf"), []byte(conf), 0o644)
}

// defaultPhpVersion picks the highest brew php@X installed, falling back to 8.2.
func defaultPhpVersion() string {
	best := ""
	entries, err := os.ReadDir(filepath.Join(brewPrefix(), "opt"))
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

func writeSiteNginxConfig(sites *Sites, name string) error {
	return writeSiteNginxConfigWithNames(sites, name, nil)
}

// writeSiteNginxConfigWithNames writes a site's nginx config. extra names are
// additional server_name values (used by `xpier share` so the tunnel host
// resolves to the site).
func writeSiteNginxConfigWithNames(sites *Sites, name string, extra []string) error {
	site, ok := sites.Sites[name]
	if !ok {
		return fmt.Errorf("site %s not linked", name)
	}
	php := site.PHP
	if php == "" {
		php = defaultPhpVersion()
	}
	domain := siteDomain(sites, name)
	root := siteRoot(site)
	cert := filepath.Join(xpierHome(), "certs", "wildcard."+sites.TLD+".pem")
	certKey := filepath.Join(xpierHome(), "certs", "wildcard."+sites.TLD+"-key.pem")
	// Prefer a per-domain cert (signed via `sudo xpier secure <domain>`) when
	// one exists; the *.test wildcard does not cover multi-label hosts like
	// img.test28.test.
	if dc, dk := domainCertPaths(domain); fileExists(dc) && fileExists(dk) {
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
	conf.WriteString("    listen 80;\n")
	conf.WriteString("    listen 443 ssl;\n")
	fmt.Fprintf(&conf, "    server_name %s;\n", serverNames)
	fmt.Fprintf(&conf, "    ssl_certificate     %s;\n", cert)
	fmt.Fprintf(&conf, "    ssl_certificate_key %s;\n", certKey)
	fmt.Fprintf(&conf, "    root %s;\n", root)
	conf.WriteString("    index index.php index.html;\n")
	if site.Driver == "hyperf" {
		port := hyperfPort(site)
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
		sock := filepath.Join(xpierHome(), "run", fmt.Sprintf("php-fpm-%s.sock", php))
		conf.WriteString("    location / {\n")
		conf.WriteString("        try_files $uri $uri/ /index.php?$query_string;\n")
		conf.WriteString("    }\n")
		conf.WriteString("    location ~ \\.php$ {\n")
		fmt.Fprintf(&conf, "        fastcgi_pass unix:%s;\n", sock)
		fmt.Fprintf(&conf, "        include %s;\n", fastcgiParamsPath())
		conf.WriteString("        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;\n")
		conf.WriteString("    }\n")
	} else {
		conf.WriteString("    location / {\n")
		conf.WriteString("        try_files $uri $uri/ =404;\n")
		conf.WriteString("    }\n")
	}
	conf.WriteString("}\n")

	if err := os.MkdirAll(nginxConfDir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(siteConfPath(name), []byte(conf.String()), 0o644)
}

func removeSiteNginxConfig(name string) error {
	err := os.Remove(siteConfPath(name))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// writeDefaultSiteConfig returns 404 for any host that is not a registered
// site, so an unlinked domain can never fall through to another site.
func writeDefaultSiteConfig() error {
	cert := filepath.Join(xpierHome(), "certs", "wildcard.test.pem")
	certKey := filepath.Join(xpierHome(), "certs", "wildcard.test-key.pem")
	conf := fmt.Sprintf(`server {
    listen 80 default_server;
    listen 443 ssl default_server;
    server_name _;
    ssl_certificate     %s;
    ssl_certificate_key %s;
    return 404;
}
`, cert, certKey)
	return os.WriteFile(filepath.Join(nginxConfDir(), "00-default.conf"), []byte(conf), 0o644)
}

// nginxReload reloads the launchd-managed nginx master via a passwordless
// sudoers entry installed by `xpier install`. The -c flag is required so
// nginx reads our pid file (defaults would target /usr/local/var/run).
func nginxReload() error {
	cmd := exec.Command("sudo", "-n", nginxBin(), "-s", "reload", "-c", filepath.Join(nginxHome(), "nginx.conf"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func loadManifestFrom(dir string) (*Manifest, error) {
	manifestPath, _ := resolvePaths(dir)
	return loadManifest(manifestPath)
}
