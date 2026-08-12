package apps

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func resolvedCwd(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return cwd
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAppConfigDevYaml(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "dev.yaml"), `namespace: devstack
apps:
  php-server:
    dir: /x
    cmd: php bin/hyperf.php server:watch
    ports: ["9501", "9502"]
`)
	cfg, cwd, err := LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Namespace != "devstack" || len(cfg.Apps) != 1 {
		t.Errorf("cfg = %+v", cfg)
	}
	if cfg.Apps["php-server"].Cmd == "" || cwd != resolvedCwd(t) {
		t.Errorf("cwd = %q, app = %+v", cwd, cfg.Apps["php-server"])
	}
}

func TestLoadAppConfigManifestApps(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "xpier.yaml"), "apps:\n  web:\n    dir: /x\n    cmd: npm run dev\n")
	cfg, _, err := LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Namespace != "default" || len(cfg.Apps) != 1 {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestLoadAppConfigMissing(t *testing.T) {
	homeTemp(t)
	chdir(t, t.TempDir())
	if _, _, err := LoadAppConfig(); err == nil {
		t.Error("LoadAppConfig without config should error")
	}
	// Empty apps in dev.yaml errors too.
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "dev.yaml"), "namespace: x\n")
	if _, _, err := LoadAppConfig(); err == nil {
		t.Error("empty apps should error")
	}
}

func TestParseForceFlag(t *testing.T) {
	force, rest := parseForceFlag([]string{"app", "--force"})
	if !force || len(rest) != 1 || rest[0] != "app" {
		t.Errorf("parseForceFlag trailing = %v %v", force, rest)
	}
	force, rest = parseForceFlag([]string{"--force", "app"})
	if !force || len(rest) != 1 {
		t.Errorf("parseForceFlag leading = %v %v", force, rest)
	}
	force, rest = parseForceFlag([]string{"app"})
	if force || len(rest) != 1 {
		t.Errorf("parseForceFlag plain = %v %v", force, rest)
	}
}

func TestAppPorts(t *testing.T) {
	app := store.App{Ports: []string{"9501", "9502"}}
	got := appPorts(app, &store.AppState{})
	if len(got) != 2 || got[0] != "9501" {
		t.Errorf("appPorts ports = %v", got)
	}
	app = store.App{Port: "8080"}
	if got := appPorts(app, &store.AppState{}); len(got) != 1 || got[0] != "8080" {
		t.Errorf("appPorts port = %v", got)
	}
	if got := appPorts(store.App{}, &store.AppState{Ports: []string{"9000"}}); len(got) != 1 || got[0] != "9000" {
		t.Errorf("appPorts state = %v", got)
	}
	if got := appPorts(store.App{}, &store.AppState{}); got != nil {
		t.Errorf("appPorts none = %v", got)
	}
}

func TestAppURL(t *testing.T) {
	if got := appURL(store.App{Domain: "abc.test"}, nil); got != "http://abc.test/" {
		t.Errorf("appURL domain = %q", got)
	}
	if got := appURL(store.App{Port: "5173"}, nil); got != "http://127.0.0.1:5173/" {
		t.Errorf("appURL port = %q", got)
	}
	if got := appURL(store.App{Ports: []string{"5173"}}, nil); got != "http://127.0.0.1:5173/" {
		t.Errorf("appURL ports = %q", got)
	}
	if got := appURL(store.App{}, &store.AppState{Port: "9000"}); got != "http://127.0.0.1:9000/" {
		t.Errorf("appURL state = %q", got)
	}
	if got := appURL(store.App{}, nil); got != "-" {
		t.Errorf("appURL none = %q", got)
	}
}

func TestDetectAppPorts(t *testing.T) {
	log := filepath.Join(t.TempDir(), "app.log")
	writeFile(t, log, "dev server running at localhost:5173\napi at 127.0.0.1:9501\n")
	got := detectAppPorts(log, []string{"8080"})
	found := map[string]bool{}
	for _, p := range got {
		found[p] = true
	}
	if !found["5173"] || !found["9501"] || len(found) != 2 {
		t.Errorf("detectAppPorts = %v", got)
	}
	if got := detectAppPorts(log, nil); len(got) != 2 {
		t.Errorf("detectAppPorts no known = %v", got)
	}
	if got := detectAppPorts(filepath.Join(t.TempDir(), "nope.log"), nil); len(got) != 0 {
		t.Errorf("detectAppPorts missing log = %v", got)
	}
}

func TestAppParseMajorAndNodeSatisfies(t *testing.T) {
	if appParseMajor("v20.11.0") != 20 || appParseMajor("20") != 20 || appParseMajor("") != 0 {
		t.Error("appParseMajor mismatch")
	}
	if out, err := exec.Command("node", "--version").Output(); err == nil && strings.TrimSpace(string(out)) != "" {
		// Installed node satisfies a tiny requirement, never an impossible one.
		if !appNodeSatisfies("0") {
			t.Error("appNodeSatisfies(0) should be true with node installed")
		}
		if appNodeSatisfies("999") {
			t.Error("appNodeSatisfies(999) should be false")
		}
	}
}

func TestSumIntsAndOrDash(t *testing.T) {
	if sumInts([]int{1, 2, 3}) != 6 || sumInts(nil) != 0 {
		t.Error("sumInts mismatch")
	}
	if orDash("x") != "x" || orDash("") != "-" {
		t.Error("orDash mismatch")
	}
}

func TestAppConfigHasDomain(t *testing.T) {
	cfg := &store.AppConfig{Apps: map[string]store.App{"a": {Domain: "x.test"}}}
	if !appConfigHasDomain(cfg) {
		t.Error("appConfigHasDomain should detect domain")
	}
	cfg = &store.AppConfig{Apps: map[string]store.App{"a": {Port: "8080"}}}
	if appConfigHasDomain(cfg) {
		t.Error("appConfigHasDomain false positive")
	}
}

func TestCmdStatusNoConfig(t *testing.T) {
	homeTemp(t)
	chdir(t, t.TempDir())
	if err := CmdStatus(nil); err == nil {
		t.Error("CmdStatus without config should error")
	}
}

func TestCmdStatusWithConfig(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "dev.yaml"), "namespace: t\napps:\n  web:\n    dir: /x\n    cmd: npm run dev\n    port: \"5173\"\n")
	if err := CmdStatus(nil); err != nil {
		t.Errorf("CmdStatus = %v", err)
	}
}

func TestCmdURLNoConfig(t *testing.T) {
	homeTemp(t)
	chdir(t, t.TempDir())
	if err := CmdURL([]string{"web"}); err == nil {
		t.Error("CmdURL without config should error")
	}
}

func TestCmdStartUnknownApp(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "dev.yaml"), "apps:\n  web:\n    dir: /x\n    cmd: npm run dev\n")
	if err := CmdStart([]string{"ghost"}); err == nil {
		t.Error("CmdStart unknown app should error")
	}
	if err := CmdLog([]string{"ghost"}); err == nil {
		t.Error("CmdLog unknown app should error")
	}
}

func TestClearAppCaches(t *testing.T) {
	homeTemp(t)
	app := store.App{Dir: t.TempDir()}
	clearAppCaches(app, false)
	// Force removes runtime/container when present (after confirmation).
	app.Dir = t.TempDir()
	writeFile(t, filepath.Join(app.Dir, "runtime", "container", "proxy.php"), "x")
	r, w, _ := os.Pipe()
	w.WriteString("y\n")
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()
	clearAppCaches(app, true)
	if store.FileExists(filepath.Join(app.Dir, "runtime", "container", "proxy.php")) {
		t.Error("force clear should remove runtime/container")
	}
}

func TestCmdUpDownNoConfig(t *testing.T) {
	homeTemp(t)
	chdir(t, t.TempDir())
	if err := CmdUp(nil); err == nil {
		t.Error("CmdUp without config should error")
	}
	if err := CmdDown(nil); err == nil {
		t.Error("CmdDown without config should error")
	}
	if err := CmdRestart(nil); err == nil {
		t.Error("CmdRestart without config should error")
	}
	if err := CmdLogsAll(nil); err == nil {
		t.Error("CmdLogsAll without config should error")
	}
}

func TestWriteAppNginxConfPorts(t *testing.T) {
	homeTemp(t)
	app := store.App{Domain: "foo.test", Ports: []string{"8080"}}
	if err := writeAppNginxConf("ns", "web", app); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(appNginxConfPath("ns", "web"))
	if err != nil {
		t.Fatal(err)
	}
	conf := string(data)
	if !strings.Contains(conf, "proxy_pass http://127.0.0.1:8080;") {
		t.Errorf("ports[0] not used:\n%s", conf)
	}
	if !strings.Contains(conf, "server_name foo.test;") {
		t.Errorf("domain missing:\n%s", conf)
	}
	// No domain -> no config written.
	if err := writeAppNginxConf("ns", "x", store.App{Port: "9000"}); err != nil {
		t.Fatal(err)
	}
	if store.FileExists(appNginxConfPath("ns", "x")) {
		t.Error("config without domain should not be written")
	}
}

func TestAppDownLiveButNotOurs(t *testing.T) {
	homeTemp(t)
	// State pointing at the test process with a marker that does not match:
	// appDown must not kill us, just drop the stale state.
	s := &store.AppState{Name: "web", PID: os.Getpid(), Cmd: "__zz_not_ours__"}
	if err := store.SaveAppState(s, "ns"); err != nil {
		t.Fatal(err)
	}
	appDown("ns", "web", store.App{})
	if !store.PidAlive(os.Getpid()) {
		t.Fatal("appDown killed the test process!")
	}
	if _, err := store.LoadAppState("ns", "web"); err == nil {
		t.Error("app state should be removed")
	}
}

func TestWriteAppNginxConfPrefersStatePort(t *testing.T) {
	homeTemp(t)
	// Config says 5173, but a running app detected 5174 (vite re-bound).
	app := store.App{Domain: "foo.test", Port: "5173"}
	s := &store.AppState{Name: "web", PID: os.Getpid(), Port: "5174", Domain: "foo.test"}
	if err := store.SaveAppState(s, "ns"); err != nil {
		t.Fatal(err)
	}
	if err := writeAppNginxConf("ns", "web", app); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(appNginxConfPath("ns", "web"))
	if !strings.Contains(string(data), "proxy_pass http://127.0.0.1:5174;") {
		t.Errorf("state port not used:\n%s", data)
	}
}

func TestProcGroupOf(t *testing.T) {
	homeTemp(t)
	if got := procGroupOf(os.Getpid()); got != syscall.Getpgrp() {
		t.Errorf("procGroupOf = %d, want %d", got, syscall.Getpgrp())
	}
	if got := procGroupOf(999999); got != 0 {
		t.Errorf("procGroupOf dead pid = %d, want 0", got)
	}
}

func TestCmdLogsAllNoRunningApps(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "dev.yaml"), "apps:\n  web:\n    dir: /x\n    cmd: npm run dev\n")
	if err := CmdLogsAll(nil); err == nil || !strings.Contains(err.Error(), "no apps running") {
		t.Errorf("CmdLogsAll no running = %v", err)
	}
}
