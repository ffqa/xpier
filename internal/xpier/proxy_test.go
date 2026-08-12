package xpier

import (
	"os"
	"path/filepath"
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

var _ = os.Getwd

func TestCmdLog(t *testing.T) {
	homeTemp(t)
	if err := cmdLog(nil); err == nil {
		t.Error("cmdLog without log should error")
	}
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdLog([]string{"nope"}); err == nil {
		t.Error("cmdLog unknown site should error")
	}
	// Unknown site checked first; a real site with a log file prints it.
	s := store.DefaultSites()
	s.Sites["mysite"] = store.Site{Path: dir, Driver: "laravel", PHP: "8.2"}
	s.Save()
	if err := os.MkdirAll(filepath.Join(store.XpierHome(), "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "php-fpm-8.2.log")
	os.WriteFile(logPath, []byte("hello log\n"), 0o644)
	if err := cmdLog([]string{"mysite"}); err != nil {
		t.Errorf("cmdLog = %v", err)
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
