package store

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func homeTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestXpierHome(t *testing.T) {
	home := homeTemp(t)
	if got := XpierHome(); got != filepath.Join(home, ".xpier") {
		t.Errorf("XpierHome = %q, want %q", got, filepath.Join(home, ".xpier"))
	}
}

func TestSlugForAndSlugName(t *testing.T) {
	homeTemp(t)
	a := SlugFor("/tmp/some/project")
	b := SlugFor("/tmp/some/project")
	if a != b {
		t.Errorf("SlugFor not deterministic: %q vs %q", a, b)
	}
	if len(a) != 8 {
		t.Errorf("SlugFor length = %d, want 8", len(a))
	}
	if SlugFor("/tmp/a") == SlugFor("/tmp/b") {
		t.Error("SlugFor collision on distinct paths")
	}
	name := SlugName("/tmp/some/project")
	if !strings.HasSuffix(name, "_"+a) || !strings.HasPrefix(name, "project_") {
		t.Errorf("SlugName = %q, want suffix _%s", name, a)
	}
}

func TestPidAlive(t *testing.T) {
	if PidAlive(0) || PidAlive(-1) {
		t.Error("PidAlive should be false for non-positive pids")
	}
	if !PidAlive(os.Getpid()) {
		t.Error("PidAlive should be true for the test process")
	}
}

func TestFileExists(t *testing.T) {
	dir := homeTemp(t)
	p := filepath.Join(dir, "f")
	if FileExists(p) {
		t.Error("FileExists true for missing file")
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(p) {
		t.Error("FileExists false for existing file")
	}
}

func TestProjectPathsAndResolvePaths(t *testing.T) {
	home := homeTemp(t)
	dir := filepath.Join(home, "proj")
	mp, lp := ProjectPaths(dir)
	if !strings.HasPrefix(mp, filepath.Join(home, ".xpier", "projects", "proj_")) {
		t.Errorf("ProjectPaths manifest = %q, want under %s", mp, filepath.Join(home, ".xpier", "projects"))
	}
	if !strings.HasSuffix(mp, ManifestName) || !strings.HasSuffix(lp, LockName) {
		t.Errorf("unexpected suffixes: %q %q", mp, lp)
	}
	// No local manifest -> project paths.
	r1, r2 := ResolvePaths(dir)
	if r1 != mp || r2 != lp {
		t.Errorf("ResolvePaths = (%q,%q), want (%q,%q)", r1, r2, mp, lp)
	}
	// Local manifest wins.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(dir, ManifestName)
	if err := os.WriteFile(local, []byte("php: 8.2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r1, r2 = ResolvePaths(dir)
	if r1 != local || r2 != filepath.Join(dir, LockName) {
		t.Errorf("ResolvePaths with local manifest = (%q,%q)", r1, r2)
	}
}

func TestEnsureProjectDir(t *testing.T) {
	home := homeTemp(t)
	dir := filepath.Join(home, "p")
	if err := EnsureProjectDir(dir); err != nil {
		t.Fatal(err)
	}
	mp, _ := ProjectPaths(dir)
	if !FileExists(filepath.Dir(mp)) {
		t.Error("project dir not created")
	}
}

func TestDefaultManifest(t *testing.T) {
	m := DefaultManifest()
	if m.Runtime != "fpm" {
		t.Errorf("default runtime = %q, want fpm", m.Runtime)
	}
}

func TestManifestRoundTrip(t *testing.T) {
	dir := homeTemp(t)
	p := filepath.Join(dir, ManifestName)
	m := &Manifest{PHP: "8.3", Runtime: "hyperf", Extensions: map[string]string{"swoole": "^5.0"}, Services: []string{"mysql"}, Apps: map[string]App{"web": {Dir: "/x", Cmd: "run", Port: "8080"}}}
	if err := m.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := LoadManifest(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.PHP != "8.3" || got.Runtime != "hyperf" || got.Services[0] != "mysql" {
		t.Errorf("round trip mismatch: %+v", got)
	}
	if got.Extensions["swoole"] != "^5.0" || got.Apps["web"].Port != "8080" {
		t.Errorf("round trip mismatch: %+v", got)
	}
	if _, err := LoadManifest(filepath.Join(dir, "missing")); err == nil {
		t.Error("LoadManifest on missing file should error")
	}
}

func TestLockRoundTrip(t *testing.T) {
	dir := homeTemp(t)
	p := filepath.Join(dir, LockName)
	l := &Lock{SchemaVersion: 1, GeneratedAt: "now", PHP: PhpLock{Version: "8.2", Path: "/bin/php"}, Extensions: []ExtLock{{Name: "redis", Constraint: "*", Installed: "5.3.7", Loaded: true}}, Services: []ServiceLock{{Name: "mysql", Running: true}}}
	if err := l.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLock(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 || got.PHP.Version != "8.2" || got.Extensions[0].Name != "redis" || !got.Services[0].Running {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestSitesRegistry(t *testing.T) {
	homeTemp(t)
	s, err := LoadSites()
	if err != nil {
		t.Fatal(err)
	}
	if s.TLD != "test" || s.Sites == nil {
		t.Errorf("default sites = %+v", s)
	}
	s.Sites["a"] = Site{Path: "/tmp/a", Driver: "laravel"}
	s.Parked = []string{"/tmp/park"}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSites()
	if err != nil {
		t.Fatal(err)
	}
	if got.TLD != "test" || got.Sites["a"].Driver != "laravel" || len(got.Parked) != 1 {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestProxiesRegistry(t *testing.T) {
	homeTemp(t)
	m, err := LoadProxies()
	if err != nil || len(m) != 0 {
		t.Fatalf("empty proxies: %v %v", m, err)
	}
	m = map[string]string{"meilisearch": "127.0.0.1:7700"}
	if err := SaveProxies(m); err != nil {
		t.Fatal(err)
	}
	got, err := LoadProxies()
	if err != nil {
		t.Fatal(err)
	}
	if got["meilisearch"] != "127.0.0.1:7700" {
		t.Errorf("round trip mismatch: %v", got)
	}
}

func TestAppState(t *testing.T) {
	homeTemp(t)
	if p := AppStatePath("ns", "web"); !strings.HasSuffix(p, "ns/web.json") {
		t.Errorf("AppStatePath = %q", p)
	}
	if p := AppLogPath("ns", "web"); !strings.HasSuffix(p, "logs/dev-web.log") {
		t.Errorf("AppLogPath = %q", p)
	}
	s := &AppState{Name: "web", PID: 42, Port: "8080"}
	if err := SaveAppState(s, "ns"); err != nil {
		t.Fatal(err)
	}
	got, err := LoadAppState("ns", "web")
	if err != nil {
		t.Fatal(err)
	}
	if got.PID != 42 || got.Port != "8080" {
		t.Errorf("round trip mismatch: %+v", got)
	}
	if _, err := LoadAppState("ns", "missing"); err == nil {
		t.Error("LoadAppState on missing should error")
	}
}

func TestSiteDomainAndRoot(t *testing.T) {
	s := &Sites{TLD: "test"}
	if SiteDomain(s, "abc") != "abc.test" {
		t.Errorf("SiteDomain = %q", SiteDomain(s, "abc"))
	}
	laravel := Site{Path: "/p", Driver: "laravel"}
	if SiteRoot(laravel) != filepath.Join("/p", "public") {
		t.Errorf("SiteRoot laravel = %q", SiteRoot(laravel))
	}
	spa := Site{Path: "/p", Driver: "spa"}
	if SiteRoot(spa) != filepath.Join("/p", "dist") {
		t.Errorf("SiteRoot spa = %q", SiteRoot(spa))
	}
	other := Site{Path: "/p", Driver: "static"}
	if SiteRoot(other) != "/p" {
		t.Errorf("SiteRoot static = %q", SiteRoot(other))
	}
}

func TestRegexes(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"abc", true}, {"my-site_1", true}, {"img.test28", true}, {".hidden", false}, {"UPPER", false}, {"a b", false}, {"a/b", false},
	}
	for _, c := range cases {
		if got := SafeSiteNameRe.MatchString(c.in); got != c.want {
			t.Errorf("SafeSiteNameRe(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	for _, v := range []string{"8.2", "8.3"} {
		if !SafePhpRe.MatchString(v) {
			t.Errorf("SafePhpRe(%q) should match", v)
		}
	}
	for _, v := range []string{"8", "8.2.1", "x.y"} {
		if SafePhpRe.MatchString(v) {
			t.Errorf("SafePhpRe(%q) should not match", v)
		}
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"b": 1, "a": 2, "c": 3}
	got := SortedKeys(m)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}

func TestDnsmasqConfig(t *testing.T) {
	homeTemp(t)
	if !strings.HasSuffix(DnsmasqConfPath(), "dnsmasq/dnsmasq.conf") {
		t.Errorf("DnsmasqConfPath = %q", DnsmasqConfPath())
	}
	if err := WriteDnsmasqConfig("test"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(DnsmasqConfPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "address=/.test/127.0.0.1") {
		t.Errorf("dnsmasq conf missing wildcard: %s", data)
	}
}

func TestUpDown(t *testing.T) {
	if UpDown(true) != "up" || UpDown(false) != "down" {
		t.Error("UpDown mismatch")
	}
}

func TestEnsureBrewPackageExisting(t *testing.T) {
	dir := homeTemp(t)
	bin := filepath.Join(dir, "cloudflared")
	if err := os.WriteFile(bin, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureBrewPackage(bin, "cloudflared", "cloudflared"); err != nil {
		t.Errorf("EnsureBrewPackage with existing bin = %v", err)
	}
}

func TestYAMLUnmarshal(t *testing.T) {
	var m Manifest
	if err := YAMLUnmarshal([]byte("php: 8.2\nruntime: fpm\n"), &m); err != nil {
		t.Fatal(err)
	}
	if m.PHP != "8.2" || m.Runtime != "fpm" {
		t.Errorf("unmarshal = %+v", m)
	}
}

func TestRunOut(t *testing.T) {
	out, err := RunOut("echo", "hello")
	if err != nil || out != "hello" {
		t.Errorf("RunOut echo = %q %v", out, err)
	}
	if err := RunOutErr("true"); err != nil {
		t.Errorf("RunOutErr true = %v", err)
	}
	if err := RunOutErr("__no_such_binary__"); err == nil {
		t.Error("RunOutErr missing binary should error")
	}
}

func TestCurrentUser(t *testing.T) {
	homeTemp(t)
	u, err := CurrentUser()
	if err != nil || u.Username == "" {
		t.Errorf("CurrentUser = %+v %v", u, err)
	}
}

func TestPortBusy(t *testing.T) {
	homeTemp(t)
	// A random high port should be free in the test environment.
	busy, err := PortBusy("59999")
	if err != nil {
		t.Fatalf("PortBusy = %v", err)
	}
	if busy {
		t.Skip("port 59999 unexpectedly busy")
	}
}

func TestKillGroup(t *testing.T) {
	homeTemp(t)
	// Killing a nonexistent group falls back to killing the single pid.
	err := KillGroup(999999, 15)
	if err != nil && err.Error() != "no such process" {
		t.Errorf("KillGroup dead pid = %v", err)
	}
}

func TestConfirmYesNoNoInput(t *testing.T) {
	homeTemp(t)
	r, w, _ := os.Pipe()
	w.WriteString("n\n")
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	ok, err := ConfirmYesNo("test?")
	if err != nil || ok {
		t.Errorf("ConfirmYesNo n = %v %v", ok, err)
	}
	r2, w2, _ := os.Pipe()
	w2.WriteString("yes\n")
	w2.Close()
	os.Stdin = r2
	ok, err = ConfirmYesNo("test?")
	if err != nil || !ok {
		t.Errorf("ConfirmYesNo yes = %v %v", ok, err)
	}
}

func TestEnsureBrewPackageMissingDeclined(t *testing.T) {
	homeTemp(t)
	bin := filepath.Join(t.TempDir(), "not-installed")
	r, w, _ := os.Pipe()
	w.WriteString("n\n")
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	err := EnsureBrewPackage(bin, "cloudflared", "cloudflared")
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("EnsureBrewPackage declined = %v", err)
	}
}

func TestPaintWord(t *testing.T) {
	green, red, yellow := "\x1b[32m", "\x1b[31m", "\x1b[33m"
	reset := "\x1b[0m"
	cases := []struct{ in, want string }{
		{"up", green + "up" + reset},
		{"up*", green + "up*" + reset},
		{"down", red + "down" + reset},
		{"none", yellow + "none" + reset},
		{"no shares", yellow + "no shares" + reset},
		{"no isolated sites", yellow + "no isolated sites" + reset},
		{"xdebug", "xdebug"},
	}
	for _, c := range cases {
		if got := paintWord(c.in); got != c.want {
			t.Errorf("paintWord(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// Paint respects NO_COLOR even on a terminal.
	t.Setenv("NO_COLOR", "1")
	if got := Paint("up"); got != "up" {
		t.Errorf("Paint with NO_COLOR = %q, want plain", got)
	}
}

func TestProcAlive(t *testing.T) {
	homeTemp(t)
	// Our own process is alive; a marker found in our cmdline matches.
	if !ProcAlive(os.Getpid(), "") {
		t.Error("ProcAlive empty marker should fall back to PidAlive")
	}
	if !ProcAlive(os.Getpid(), "store.test") {
		t.Error("ProcAlive should match our test binary cmdline")
	}
	if ProcAlive(os.Getpid(), "__zz_no_such_marker__") {
		t.Error("ProcAlive should reject a marker absent from the cmdline")
	}
	if ProcAlive(999999, "anything") {
		t.Error("ProcAlive should be false for a dead pid")
	}
}

func TestResolvePathsLegacyManifest(t *testing.T) {
	homeTemp(t)
	dir := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Old-style local manifest is still detected.
	legacy := filepath.Join(dir, LegacyManifestName)
	if err := os.WriteFile(legacy, []byte("php: 8.2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mp, _ := ResolvePaths(dir)
	if mp != legacy {
		t.Errorf("legacy manifest not detected: %q", mp)
	}
}
