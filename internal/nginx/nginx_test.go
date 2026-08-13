package nginx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestCompareVersionStrings(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"8.2", "8.2", 0},
		{"8.3", "8.2", 1},
		{"8.2", "8.3", -1},
		{"8.2.1", "8.2.0", 1},
		{"8.10", "8.9", 1},
		{"8", "8.0", 0},
	}
	for _, c := range cases {
		if got := compareVersionStrings(c.a, c.b); got != c.want {
			t.Errorf("compareVersionStrings(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestIsVersionString(t *testing.T) {
	for _, ok := range []string{"8.2", "8.2.31", "8.10"} {
		if !isVersionString(ok) {
			t.Errorf("isVersionString(%q) should be true", ok)
		}
	}
	for _, no := range []string{"", "php", "8", "8.", "8.x", "8.2.1-beta"} {
		if isVersionString(no) {
			t.Errorf("isVersionString(%q) should be false", no)
		}
	}
}

func TestDefaultPhpVersion(t *testing.T) {
	homeTemp(t)
	v := DefaultPhpVersion()
	if !strings.Contains(v, ".") {
		t.Errorf("DefaultPhpVersion = %q, want x.y", v)
	}
}

func TestPathHelpers(t *testing.T) {
	homeTemp(t)
	if !strings.HasSuffix(NginxHome(), "/.xpier/nginx") {
		t.Errorf("NginxHome = %q", NginxHome())
	}
	if !strings.HasSuffix(NginxConfDir(), "/nginx/conf.d") {
		t.Errorf("NginxConfDir = %q", NginxConfDir())
	}
	if SiteConfPath("abc") != filepath.Join(NginxConfDir(), "abc.conf") {
		t.Errorf("SiteConfPath = %q", SiteConfPath("abc"))
	}
	if !strings.Contains(FastcgiParamsPath(), "fastcgi_params") {
		t.Errorf("FastcgiParamsPath = %q", FastcgiParamsPath())
	}
	c, k := CertPaths("test")
	if !strings.HasSuffix(c, "certs/wildcard.test.pem") || !strings.HasSuffix(k, "certs/wildcard.test-key.pem") {
		t.Errorf("CertPaths = %q %q", c, k)
	}
	dc, dk := DomainCertPaths("img.test28.test")
	if !strings.HasSuffix(dc, "certs/img.test28.test.pem") || !strings.HasSuffix(dk, "certs/img.test28.test-key.pem") {
		t.Errorf("DomainCertPaths = %q %q", dc, dk)
	}
}

func TestWriteNginxMainConfig(t *testing.T) {
	homeTemp(t)
	if err := WriteNginxMainConfig(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(NginxHome(), "nginx.conf"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"worker_processes auto;", "server_names_hash_bucket_size 256;", "include " + NginxConfDir() + "/*.conf;"} {
		if !strings.Contains(s, want) {
			t.Errorf("nginx.conf missing %q", want)
		}
	}
}

func siteSet(t *testing.T) *store.Sites {
	t.Helper()
	return &store.Sites{TLD: "test", Sites: map[string]store.Site{
		"larablog": {Path: "/srv/larablog", Driver: "laravel", PHP: "8.2"},
		"hyperf":   {Path: "/srv/hyperf", Driver: "hyperf"},
		"static":   {Path: "/srv/static", Driver: "static"},
	}}
}

func TestWriteSiteNginxConfigLaravel(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfig(s, "larablog"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(SiteConfPath("larablog"))
	conf := string(data)
	for _, want := range []string{
		"server_name larablog.test;",
		"root \"/srv/larablog/public\";",
		"fastcgi_pass unix:" + filepath.Join(store.XpierHome(), "run", "php-fpm-8.2.sock") + ";",
		"try_files $uri $uri/ /index.php?$query_string;",
		"ssl_certificate",
	} {
		if !strings.Contains(conf, want) {
			t.Errorf("laravel conf missing %q:\n%s", want, conf)
		}
	}
}

func TestWriteSiteNginxConfigHyperf(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfig(s, "hyperf"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(SiteConfPath("hyperf"))
	conf := string(data)
	for _, want := range []string{
		"proxy_pass http://127.0.0.1:9501;",
		`proxy_set_header Upgrade $http_upgrade;`,
		`proxy_set_header Connection "upgrade";`,
	} {
		if !strings.Contains(conf, want) {
			t.Errorf("hyperf conf missing %q:\n%s", want, conf)
		}
	}
}

func TestWriteSiteNginxConfigStatic(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfig(s, "static"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(SiteConfPath("static"))
	conf := string(data)
	if !strings.Contains(conf, "try_files $uri $uri/ =404;") {
		t.Errorf("static conf missing 404 try_files:\n%s", conf)
	}
}

func TestWriteSiteNginxConfigExtraNames(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfigWithNames(s, "larablog", []string{"abc.trycloudflare.com"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(SiteConfPath("larablog"))
	if !strings.Contains(string(data), "server_name larablog.test abc.trycloudflare.com;") {
		t.Errorf("extra server names not written:\n%s", data)
	}
}

func TestWriteSiteNginxConfigUnknownSite(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfig(s, "nope"); err == nil {
		t.Error("unknown site should error")
	}
}

func TestRemoveSiteNginxConfig(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteSiteNginxConfig(s, "larablog"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveSiteNginxConfig("larablog"); err != nil {
		t.Fatal(err)
	}
	if store.FileExists(SiteConfPath("larablog")) {
		t.Error("config not removed")
	}
	if err := RemoveSiteNginxConfig("never-existed"); err != nil {
		t.Errorf("removing missing config = %v", err)
	}
}

func TestWriteDefaultSiteConfig(t *testing.T) {
	homeTemp(t)
	if err := WriteDefaultSiteConfig(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(NginxConfDir(), "00-default.conf"))
	if !strings.Contains(string(data), "server_name _;") || !strings.Contains(string(data), "return 404;") {
		t.Errorf("default conf wrong:\n%s", data)
	}
}

func TestWriteAllSiteConfigs(t *testing.T) {
	homeTemp(t)
	s := siteSet(t)
	if err := WriteAllSiteConfigs(s); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"larablog", "hyperf", "static"} {
		if !store.FileExists(SiteConfPath(name)) {
			t.Errorf("config for %s not written", name)
		}
	}
}

func TestServerPortsAndHyperfPort(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "config", "autoload"), 0o755); err != nil {
		t.Fatal(err)
	}
	php := `<?php
return [
    'servers' => [
        'http' => [
            'name' => 'http',
            'host' => '0.0.0.0',
            'port' => 9501,
        ],
        'ws' => [
            'name' => 'ws',
            'host' => '0.0.0.0',
            'port' => 9502,
        ],
    ],
];`
	if err := os.WriteFile(filepath.Join(dir, "config", "autoload", "server.php"), []byte(php), 0o644); err != nil {
		t.Fatal(err)
	}
	ports := ServerPorts(dir)
	if ports["http"] != "9501" || ports["ws"] != "9502" {
		t.Errorf("ServerPorts = %v", ports)
	}
	if HyperfPort(store.Site{Path: dir}) != "9501" {
		t.Errorf("HyperfPort = %q", HyperfPort(store.Site{Path: dir}))
	}
	if HyperfPort(store.Site{Path: t.TempDir()}) != "9501" {
		t.Error("HyperfPort fallback should be 9501")
	}
}

func TestCurrentTLD(t *testing.T) {
	homeTemp(t)
	if got := CurrentTLD(); got != "test" {
		t.Errorf("CurrentTLD default = %q, want test", got)
	}
	s := store.DefaultSites()
	s.TLD = "dev"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if got := CurrentTLD(); got != "dev" {
		t.Errorf("CurrentTLD = %q, want dev", got)
	}
}

func TestWriteDefaultSiteConfigUsesCurrentTLD(t *testing.T) {
	homeTemp(t)
	s := store.DefaultSites()
	s.TLD = "dev"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if err := WriteDefaultSiteConfig(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(NginxConfDir(), "00-default.conf"))
	if err != nil {
		t.Fatal(err)
	}
	conf := string(data)
	if !strings.Contains(conf, "wildcard.dev.pem") || strings.Contains(conf, "wildcard.test.pem") {
		t.Errorf("default site cert should reference current TLD:\n%s", conf)
	}
}

func TestDefaultPhpVersionPinned(t *testing.T) {
	homeTemp(t)
	s := store.DefaultSites()
	s.DefaultPHP = "7.4"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if got := DefaultPhpVersion(); got != "7.4" {
		t.Errorf("pinned default = %q, want 7.4", got)
	}
}

func TestWriteSiteNginxConfigUnsecure(t *testing.T) {
	homeTemp(t)
	f := false
	s := &store.Sites{TLD: "test", Sites: map[string]store.Site{
		"http": {Path: "/srv/http", Driver: "laravel", PHP: "8.2", Secure: &f},
	}}
	if err := WriteSiteNginxConfig(s, "http"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(SiteConfPath("http"))
	conf := string(data)
	if strings.Contains(conf, "443") || strings.Contains(conf, "ssl_certificate") {
		t.Errorf("unsecure site should have no TLS:\n%s", conf)
	}
	if !strings.Contains(conf, "listen 127.0.0.1:80;") {
		t.Errorf("unsecure site missing listen 80:\n%s", conf)
	}
	// Default (Secure nil) still has TLS.
	s2 := siteSet(t)
	if err := WriteSiteNginxConfig(s2, "larablog"); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(SiteConfPath("larablog"))
	if !strings.Contains(string(data), "listen 127.0.0.1:443 ssl;") {
		t.Error("default site should keep TLS")
	}
}
