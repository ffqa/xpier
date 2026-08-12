package sites

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"xpier/internal/service"
	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(old) })
}

func TestDetectDriver(t *testing.T) {
	base := t.TempDir()
	laravel := filepath.Join(base, "laravel")
	hyperf := filepath.Join(base, "hyperf")
	spa := filepath.Join(base, "spa")
	static := filepath.Join(base, "static")
	for _, d := range []string{laravel, hyperf, spa, static} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(t, filepath.Join(laravel, "public", "index.php"), "<?php")
	mustWrite(t, filepath.Join(hyperf, "bin", "hyperf.php"), "<?php")
	mustWrite(t, filepath.Join(spa, "dist", "index.html"), "<html>")
	cases := []struct {
		dir, want string
	}{
		{laravel, "laravel"},
		{hyperf, "hyperf"},
		{spa, "spa"},
		{static, "static"},
	}
	for _, c := range cases {
		if got := DetectDriver(c.dir); got != c.want {
			t.Errorf("DetectDriver(%s) = %q, want %q", c.dir, got, c.want)
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSite(t *testing.T) {
	homeTemp(t)
	s := &store.Sites{TLD: "test", Sites: map[string]store.Site{
		"abc": {Path: "/srv/abc", Driver: "laravel"},
	}}
	// Explicit name.
	name, site, err := ResolveSite(s, "abc")
	if err != nil || name != "abc" || site.Path != "/srv/abc" {
		t.Errorf("ResolveSite explicit = %q %+v %v", name, site, err)
	}
	// Unknown explicit name.
	if _, _, err := ResolveSite(s, "zzz"); err == nil {
		t.Error("unknown site should error")
	}
	// Current-directory match by basename.
	dir := filepath.Join(t.TempDir(), "abc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	name, _, err = ResolveSite(s, "")
	if err != nil || name != "abc" {
		t.Errorf("ResolveSite by cwd = %q %v", name, err)
	}
	// Match by resolved path when the key differs from basename.
	cwd, _ := os.Getwd()
	s2 := &store.Sites{TLD: "test", Sites: map[string]store.Site{
		"custom": {Path: cwd, Driver: "laravel"},
	}}
	name, _, err = ResolveSite(s2, "")
	if err != nil || name != "custom" {
		t.Errorf("ResolveSite by path = %q %v", name, err)
	}
	// No match anywhere.
	chdir(t, t.TempDir())
	if _, _, err := ResolveSite(s2, ""); err == nil {
		t.Error("unlinked cwd should error")
	}
}

func TestExtractSiteFlag(t *testing.T) {
	cases := []struct {
		args []string
		site string
		rest []string
	}{
		{[]string{"--site", "abc", "-r", "echo 1"}, "abc", []string{"-r", "echo 1"}},
		{[]string{"--site=abc", "hello"}, "abc", []string{"hello"}},
		{[]string{"-r", "echo 1"}, "", []string{"-r", "echo 1"}},
		{[]string{}, "", []string{}},
	}
	for _, c := range cases {
		site, rest := ExtractSiteFlag(c.args)
		if site != c.site || strings.Join(rest, " ") != strings.Join(c.rest, " ") {
			t.Errorf("ExtractSiteFlag(%v) = (%q,%v), want (%q,%v)", c.args, site, rest, c.site, c.rest)
		}
	}
}

func TestSitePHPBinUnavailable(t *testing.T) {
	homeTemp(t)
	_, _, err := SitePHPBin(store.Site{PHP: "99.9"})
	if err == nil || !strings.Contains(err.Error(), "php@99.9") {
		t.Errorf("SitePHPBin unavailable = %v", err)
	}
}

func TestSyncParked(t *testing.T) {
	homeTemp(t)
	park := filepath.Join(t.TempDir(), "parked")
	sub1 := filepath.Join(park, "one")
	sub2 := filepath.Join(park, "two")
	if err := os.MkdirAll(sub1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sub2, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(sub1, "public", "index.php"), "<?php")
	s := &store.Sites{TLD: "test", Parked: []string{park}, Sites: map[string]store.Site{}}
	s.Sites["existing"] = store.Site{Path: "/keep", Driver: "static"}
	SyncParked(s)
	if s.Sites["one"].Driver != "laravel" {
		t.Errorf("parked site one driver = %q", s.Sites["one"].Driver)
	}
	if s.Sites["two"].Driver != "static" {
		t.Errorf("parked site two driver = %q", s.Sites["two"].Driver)
	}
	if _, ok := s.Sites["existing"]; !ok {
		t.Error("existing site should be preserved")
	}
}

func TestCmdParkedEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdParked(nil); err != nil {
		t.Errorf("CmdParked empty = %v", err)
	}
}

func TestCmdLinksEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdLinks(nil); err != nil {
		t.Errorf("CmdLinks empty = %v", err)
	}
}

func regSite(t *testing.T, name string, s store.Site) {
	t.Helper()
	all := store.DefaultSites()
	all.Sites[name] = s
	if err := all.Save(); err != nil {
		t.Fatal(err)
	}
}

func TestCmdSitesEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdSites(nil); err != nil {
		t.Errorf("CmdSites empty = %v", err)
	}
}

func TestCmdSitesUpNoSites(t *testing.T) {
	homeTemp(t)
	if err := CmdSitesUp(nil); err == nil {
		t.Error("CmdSitesUp with no sites should error")
	}
}

func TestCmdSitesDownNoState(t *testing.T) {
	homeTemp(t)
	if err := CmdSitesDown(nil); err == nil {
		t.Error("CmdSitesDown with no state should error")
	}
}

func TestCmdLinkRegistersSite(t *testing.T) {
	homeTemp(t)
	dir := filepath.Join(t.TempDir(), "mysite")
	mustWrite(t, filepath.Join(dir, "public", "index.php"), "<?php")
	chdir(t, dir)
	if err := CmdLink(nil); err != nil {
		t.Fatalf("CmdLink = %v", err)
	}
	s, err := store.LoadSites()
	if err != nil {
		t.Fatal(err)
	}
	site, ok := s.Sites["mysite"]
	if !ok || site.Driver != "laravel" {
		t.Errorf("site not registered: %+v", s.Sites)
	}
}

func TestCmdUnlinkNonLinked(t *testing.T) {
	homeTemp(t)
	chdir(t, t.TempDir())
	if err := CmdUnlink([]string{"nope"}); err == nil {
		t.Error("unlink non-linked should error")
	}
}

func TestCmdSitePHP(t *testing.T) {
	homeTemp(t)
	regSite(t, "abc", store.Site{Path: "/srv/abc", Driver: "laravel"})
	if err := CmdSitePHP([]string{"abc"}); err != nil {
		t.Errorf("CmdSitePHP get = %v", err)
	}
	if err := CmdSitePHP([]string{"abc", "8.3"}); err != nil {
		t.Errorf("CmdSitePHP set = %v", err)
	}
	s, _ := store.LoadSites()
	if s.Sites["abc"].PHP != "8.3" {
		t.Errorf("site php = %q, want 8.3", s.Sites["abc"].PHP)
	}
	if err := CmdSitePHP([]string{"zzz"}); err == nil {
		t.Error("unknown site should error")
	}
	if err := CmdSitePHP([]string{}); err == nil {
		t.Error("missing args should error")
	}
}

func TestCmdIsolateUnisolate(t *testing.T) {
	homeTemp(t)
	regSite(t, "abc", store.Site{Path: "/srv/abc", Driver: "laravel"})
	if err := CmdIsolate([]string{"--site", "abc", "8.3"}); err != nil {
		t.Fatalf("CmdIsolate = %v", err)
	}
	s, _ := store.LoadSites()
	if s.Sites["abc"].PHP != "8.3" {
		t.Errorf("isolate php = %q", s.Sites["abc"].PHP)
	}
	if err := CmdIsolate([]string{"--site", "abc", "9"}); err == nil {
		t.Error("invalid php should error")
	}
	if err := CmdUnisolate([]string{"--site", "abc"}); err != nil {
		t.Fatalf("CmdUnisolate = %v", err)
	}
	s, _ = store.LoadSites()
	if s.Sites["abc"].PHP != "" {
		t.Errorf("unisolate php = %q, want empty", s.Sites["abc"].PHP)
	}
	if err := CmdIsolated(nil); err != nil {
		t.Errorf("CmdIsolated = %v", err)
	}
}

func TestCmdTLDAndLoopback(t *testing.T) {
	homeTemp(t)
	if err := CmdTLD(nil); err != nil {
		t.Errorf("CmdTLD get = %v", err)
	}
	if err := CmdTLD([]string{"dev"}); err != nil {
		t.Errorf("CmdTLD set = %v", err)
	}
	s, _ := store.LoadSites()
	if s.TLD != "dev" {
		t.Errorf("tld = %q, want dev", s.TLD)
	}
	if err := CmdTLD([]string{"bad tld"}); err == nil {
		t.Error("invalid tld should error")
	}
	if err := CmdLoopback(nil); err != nil {
		t.Errorf("CmdLoopback get = %v", err)
	}
	if err := CmdLoopback([]string{"1.2.3.4"}); err != nil {
		t.Errorf("CmdLoopback set = %v", err)
	}
}

func TestCmdSiteInformation(t *testing.T) {
	homeTemp(t)
	regSite(t, "abc", store.Site{Path: "/srv/abc", Driver: "laravel", PHP: "8.2"})
	if err := CmdSiteInformation([]string{"abc"}); err != nil {
		t.Errorf("CmdSiteInformation = %v", err)
	}
	if err := CmdSiteInformation([]string{"zzz"}); err == nil {
		t.Error("unknown site should error")
	}
}

func TestCmdPark(t *testing.T) {
	homeTemp(t)
	park := t.TempDir()
	sub := filepath.Join(park, "proj")
	mustWrite(t, filepath.Join(sub, "public", "index.php"), "<?php")
	if err := CmdPark([]string{park}); err != nil {
		t.Fatalf("CmdPark = %v", err)
	}
	s, _ := store.LoadSites()
	if s.Sites["proj"].Driver != "laravel" {
		t.Errorf("parked site = %+v", s.Sites)
	}
	if err := CmdPark([]string{filepath.Join(park, "missing")}); err == nil {
		t.Error("park missing dir should error")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	data, _ := io.ReadAll(r)
	return string(data)
}

func dnsmasqStatus(s string) string {
	i := strings.Index(s, "dnsmasq:")
	if i < 0 {
		return ""
	}
	fields := strings.Fields(s[i+len("dnsmasq:"):])
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// TestDnsmasqStatusConsistent guards against the sites/services views of the
// same daemon drifting apart (they used lsof vs pgrep detection before).
func TestDnsmasqStatusConsistent(t *testing.T) {
	homeTemp(t)
	sitesOut := captureStdout(t, func() { CmdSites(nil) })
	svcOut := captureStdout(t, func() { service.XpierServiceStatus() })
	s1 := dnsmasqStatus(sitesOut)
	s2 := dnsmasqStatus(svcOut)
	if s1 == "" || s2 == "" {
		t.Fatalf("could not parse dnsmasq status: sites=%q services=%q", sitesOut, svcOut)
	}
	if s1 != s2 {
		t.Errorf("dnsmasq status inconsistent: sites=%q services=%q", s1, s2)
	}
}
