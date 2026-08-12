package xpier

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xpier/internal/store"
)

func TestCmdProxyLifecycle(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)

	if err := cmdProxy([]string{"meili"}); err == nil {
		t.Error("missing upstream should error")
	}
	if err := cmdProxy([]string{"meili", "127.0.0.1:7700"}); err != nil {
		t.Fatalf("cmdProxy = %v", err)
	}
	proxies, err := store.LoadProxies()
	if err != nil {
		t.Fatal(err)
	}
	if proxies["meili.test"] != "127.0.0.1:7700" {
		t.Errorf("proxies = %v", proxies)
	}
	if err := cmdProxies(nil); err != nil {
		t.Errorf("cmdProxies = %v", err)
	}
	// Injection attempt must be rejected.
	if err := cmdProxy([]string{"evil", "x;return 500;"}); err == nil {
		t.Error("injection upstream should error")
	}
	// Full URL upstream.
	if err := cmdProxy([]string{"api.test", "http://127.0.0.1:8000"}); err != nil {
		t.Errorf("cmdProxy url = %v", err)
	}
	// Unproxy.
	if err := cmdUnproxy([]string{"meili"}); err != nil {
		t.Errorf("cmdUnproxy = %v", err)
	}
	if err := cmdUnproxy([]string{"meili"}); err == nil {
		t.Error("second unproxy should error")
	}
	if err := cmdUnproxy([]string{}); err == nil {
		t.Error("unproxy without domain should error")
	}
}

func TestCmdProxiesEmpty(t *testing.T) {
	homeTemp(t)
	if err := cmdProxies(nil); err != nil {
		t.Errorf("cmdProxies empty = %v", err)
	}
}

func TestProxyConfPath(t *testing.T) {
	homeTemp(t)
	if !filepath.IsAbs(proxyConfPath("api.test")) {
		t.Errorf("proxyConfPath = %q", proxyConfPath("api.test"))
	}
}

func TestCmdDoctor(t *testing.T) {
	homeTemp(t)
	if err := cmdDoctor(nil); err != nil {
		t.Errorf("cmdDoctor = %v", err)
	}
}

func TestCmdStatusInitProject(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdInit([]string{"--php", "8.2"}); err != nil {
		t.Fatal(err)
	}
	if err := cmdStatus(nil); err != nil {
		t.Errorf("cmdStatus = %v", err)
	}
}

func TestCmdDirectoryListing(t *testing.T) {
	homeTemp(t)
	if err := cmdDirectoryListing([]string{"on"}); err == nil {
		t.Error("directory-listing without nginx.conf should error")
	}
}

func TestCmdXdebug(t *testing.T) {
	homeTemp(t)
	if err := cmdXdebug([]string{"status"}); err != nil {
		t.Errorf("cmdXdebug status = %v", err)
	}
	if err := cmdXdebug([]string{"off"}); err != nil {
		t.Errorf("cmdXdebug off = %v", err)
	}
	if err := cmdXdebug([]string{"bogus"}); err == nil {
		t.Error("cmdXdebug bad arg should error")
	}
}

func TestCmdMailDownNoState(t *testing.T) {
	homeTemp(t)
	if err := cmdMailDown(nil); err == nil {
		t.Error("cmdMailDown without state should error")
	}
}

func TestMailStateRoundTrip(t *testing.T) {
	homeTemp(t)
	if err := writeMailState(123, "/tmp/l"); err != nil {
		t.Fatal(err)
	}
	st, err := loadMailState()
	if err != nil {
		t.Fatal(err)
	}
	if st.PID != 123 || st.LogPath != "/tmp/l" {
		t.Errorf("mail state = %+v", st)
	}
}

func TestPlanExtMissingWithPHP(t *testing.T) {
	homeTemp(t)
	bin := phpBinFor("8.2")
	if v := phpVersion(bin); v == "" {
		t.Skip("php@8.2 not installed on this machine")
	}
	m := &store.Manifest{PHP: "8.2", Extensions: map[string]string{"__zz_not_a_real_ext__": "^1.0"}}
	items := plan(m)
	for _, it := range items {
		if it.kind == "ext" && it.name == "__zz_not_a_real_ext__" {
			if it.state != "missing" {
				t.Errorf("fake ext should be missing, got %s", it.state)
			}
			return
		}
	}
	t.Error("fake ext item not found in plan")
}

func TestCmdIniMissing(t *testing.T) {
	homeTemp(t)
	if err := cmdIni([]string{"--php", "99.9"}); err == nil {
		t.Error("cmdIni missing php.ini should error")
	}
}

func TestAdminerURL(t *testing.T) {
	homeTemp(t)
	s := store.DefaultSites()
	if got := adminerURL(s, ""); got != "http://database.test/" {
		t.Errorf("adminerURL no site = %q", got)
	}
	if got := adminerURL(s, "larablog"); got != "http://database.test/?db=larablog" {
		t.Errorf("adminerURL with site = %q", got)
	}
	s.TLD = "dev"
	if got := adminerURL(s, ""); got != "http://database.dev/" {
		t.Errorf("adminerURL custom tld = %q", got)
	}
}

func TestEnsureAdminerSiteEmbedded(t *testing.T) {
	homeTemp(t)
	if err := ensureAdminerSite(); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(store.XpierHome(), "adminer", "index.php")
	if !store.FileExists(index) {
		t.Fatal("adminer index.php not written from embedded copy")
	}
	data, err := os.ReadFile(index)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || !strings.Contains(string(data[:200]), "Adminer") {
		t.Error("embedded adminer content missing")
	}
	// The reserved site is registered.
	s, err := store.LoadSites()
	if err != nil {
		t.Fatal(err)
	}
	if s.Sites["database"].Path != filepath.Join(store.XpierHome(), "adminer") {
		t.Errorf("database site = %+v", s.Sites["database"])
	}
	// A stale deployed copy is replaced by the embedded (patched) one.
	stale := filepath.Join(store.XpierHome(), "adminer", "index.php")
	os.WriteFile(stale, []byte("<?php // stale"), 0o644)
	if err := ensureAdminerSite(); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(stale)
	if bytes.Contains(data, []byte("stale")) {
		t.Error("stale adminer copy was not synced to the embedded version")
	}
	if !bytes.Contains(data, []byte("xpier patch")) {
		t.Error("deployed adminer is missing the empty-password patch")
	}
	// A user site named database is not overwritten.
	reg := store.DefaultSites()
	reg.Sites["database"] = store.Site{Path: "/user/db", Driver: "laravel"}
	reg.Save()
	if err := ensureAdminerSite(); err == nil {
		t.Error("ensureAdminerSite should refuse to overwrite a user database site")
	}
}
