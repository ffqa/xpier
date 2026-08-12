package xpier

import (
	"os"
	"path/filepath"
	"testing"

	"xpier/internal/store"
)

func TestPlanMissingPHP(t *testing.T) {
	homeTemp(t)
	m := &store.Manifest{PHP: "99.9", Extensions: map[string]string{"swoole": "^5.0"}, Services: []string{"__nonexistent_svc__"}}
	items := plan(m)
	found := map[string]bool{}
	for _, it := range items {
		found[it.kind+"-"+it.name] = true
		if it.kind == "php" && it.state != "missing" {
			t.Errorf("php 99.9 should be missing, got %s", it.state)
		}
		if it.kind == "ext" && it.state != "missing" {
			t.Errorf("ext without php should be missing, got %s", it.state)
		}
	}
	if !found["php-99.9"] || !found["ext-swoole"] || !found["svc-__nonexistent_svc__"] {
		t.Errorf("plan items incomplete: %v", found)
	}
}

func TestPlanInstalledPHP(t *testing.T) {
	homeTemp(t)
	bin := phpBinFor("8.2")
	if v := phpVersion(bin); v == "" {
		t.Skip("php@8.2 not installed on this machine")
	}
	m := &store.Manifest{PHP: "8.2"}
	items := plan(m)
	for _, it := range items {
		if it.kind == "php" && it.state != "ok" {
			t.Errorf("installed php should be ok, got %s (%s)", it.state, it.detail)
		}
	}
}

func TestWriteLock(t *testing.T) {
	homeTemp(t)
	bin := phpBinFor("8.2")
	if v := phpVersion(bin); v == "" {
		t.Skip("php@8.2 not installed on this machine")
	}
	lockPath := filepath.Join(t.TempDir(), "xpier.lock")
	m := &store.Manifest{PHP: "8.2", Extensions: map[string]string{}, Services: []string{}}
	if err := writeLock(m, lockPath); err != nil {
		t.Fatal(err)
	}
	lock, err := store.LoadLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if lock.PHP.Version == "" || lock.SchemaVersion != 1 {
		t.Errorf("lock = %+v", lock)
	}
}

func TestCmdInit(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdInit(nil); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	mp, _ := store.ProjectPaths(cwd)
	if !store.FileExists(mp) {
		t.Fatalf("manifest not written at %s", mp)
	}
	m, err := store.LoadManifest(mp)
	if err != nil || m.Runtime != "fpm" {
		t.Errorf("manifest = %+v %v", m, err)
	}
	// Re-init should fail (exists).
	if err := cmdInit(nil); err == nil {
		t.Error("second init should error")
	}
}

func TestCmdInitLocal(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdInit([]string{"."}); err != nil {
		t.Fatal(err)
	}
	if !store.FileExists(filepath.Join(dir, store.ManifestName)) {
		t.Error("local xpier.yaml not written")
	}
}

func TestCmdInitInvalidRuntime(t *testing.T) {
	homeTemp(t)
	if err := cmdInit([]string{"--runtime", "bogus"}); err == nil {
		t.Error("invalid runtime should error")
	}
}

func TestSyncDryRunMissingPHP(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	cwd, _ := os.Getwd()
	mp, _ := store.ProjectPaths(cwd)
	if err := os.MkdirAll(filepath.Dir(mp), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := (&store.Manifest{PHP: "99.9"}).Save(mp); err != nil {
		t.Fatal(err)
	}
	if err := cmdSync([]string{}); err != nil {
		t.Errorf("cmdSync dry-run = %v", err)
	}
}

func TestSafeExtRe(t *testing.T) {
	if !safeExtRe.MatchString("swoole") || !safeExtRe.MatchString("redis_5") {
		t.Error("safeExtRe should accept simple names")
	}
	for _, bad := range []string{"a;b", "../x", "a b"} {
		if safeExtRe.MatchString(bad) {
			t.Errorf("safeExtRe should reject %q", bad)
		}
	}
}

func TestInitManifestTemplateComplete(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdInit(nil); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	mp, _ := store.ProjectPaths(cwd)
	data, err := os.ReadFile(mp)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"runtime: fpm", "# php:", "# extensions:", "# services:", "# apps:", "swoole", "xdebug", "node:"} {
		if !strings.Contains(content, want) {
			t.Errorf("template missing %q:\n%s", want, content)
		}
	}
	// Still a valid manifest.
	m, err := store.LoadManifest(mp)
	if err != nil || m.Runtime != "fpm" {
		t.Errorf("template not parseable: %+v %v", m, err)
	}
}

func TestInitManifestTemplatePinsPHP(t *testing.T) {
	homeTemp(t)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(old)
	if err := cmdInit([]string{"--php", "8.4"}); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	mp, _ := store.ProjectPaths(cwd)
	data, _ := os.ReadFile(mp)
	if !strings.Contains(string(data), `php: "8.4"`) {
		t.Errorf("pinned php missing:\n%s", data)
	}
}
