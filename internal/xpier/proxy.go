package xpier

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

// proxy registry: domain -> upstream (host:port or full URL). For services
// that run outside xpier (meilisearch, docker containers, ...) - xpier only
// writes the nginx reverse proxy.

func proxyConfPath(domain string) string {
	return filepath.Join(nginx.NginxConfDir(), "proxy-"+domain+".conf")
}

func writeProxyConf(domain, upstream string) error {
	if !strings.HasPrefix(upstream, "http://") && !strings.HasPrefix(upstream, "https://") {
		upstream = "http://" + upstream
	}
	conf := fmt.Sprintf(`server {
    listen 80;
    listen 443 ssl;
    server_name %s;
    ssl_certificate     %s;
    ssl_certificate_key %s;
    location / {
        proxy_pass %s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`, domain, filepath.Join(store.XpierHome(), "certs", "wildcard.test.pem"),
		filepath.Join(store.XpierHome(), "certs", "wildcard.test-key.pem"), upstream)
	if err := os.MkdirAll(filepath.Dir(proxyConfPath(domain)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(proxyConfPath(domain), []byte(conf), 0o644)
}

func cmdProxy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: xpier proxy <domain> <host[:port]|http://host:port>")
	}
	domain := strings.TrimPrefix(args[0], ".")
	if !store.SafeSiteNameRe.MatchString(domain) {
		return fmt.Errorf("invalid domain %q", domain)
	}
	// Append the TLD when a bare name was given (proxy -> proxy.test).
	if sites, err := store.LoadSites(); err == nil && !strings.Contains(domain, ".") {
		domain = domain + "." + sites.TLD
	}
	// Upstream is embedded into an nginx config; reject anything that could
	// terminate the proxy_pass directive and inject other directives.
	upstream := args[1]
	if strings.ContainsAny(upstream, ";\n#{}") {
		return fmt.Errorf("invalid upstream %q (must be host[:port] or http(s)://host:port)", upstream)
	}
	proxies, err := store.LoadProxies()
	if err != nil {
		return err
	}
	proxies[domain] = upstream
	if err := store.SaveProxies(proxies); err != nil {
		return err
	}
	if err := writeProxyConf(domain, upstream); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	fmt.Printf("proxy %s -> %s\n", domain, upstream)
	return nil
}

func cmdProxies(args []string) error {
	proxies, err := store.LoadProxies()
	if err != nil {
		return err
	}
	if len(proxies) == 0 {
		fmt.Println("no proxies")
		return nil
	}
	keys := make([]string, 0, len(proxies))
	for d := range proxies {
		keys = append(keys, d)
	}
	sort.Strings(keys)
	for _, d := range keys {
		fmt.Printf("  %-28s -> %s\n", d, proxies[d])
	}
	return nil
}

func cmdUnproxy(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xpier unproxy <domain>")
	}
	domain := strings.TrimPrefix(args[0], ".")
	if sites, err := store.LoadSites(); err == nil && !strings.Contains(domain, ".") {
		domain = domain + "." + sites.TLD
	}
	proxies, err := store.LoadProxies()
	if err != nil {
		return err
	}
	if _, ok := proxies[domain]; !ok {
		return fmt.Errorf("proxy %s not found", domain)
	}
	delete(proxies, domain)
	if err := store.SaveProxies(proxies); err != nil {
		return err
	}
	os.Remove(proxyConfPath(domain))
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	fmt.Printf("unproxied %s\n", domain)
	return nil
}
