package xpier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeBin(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "fakebin")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestPhpBinFor(t *testing.T) {
	if !strings.Contains(phpBinFor("8.2"), "php@8.2/bin/php") {
		t.Errorf("phpBinFor = %q", phpBinFor("8.2"))
	}
	cands := phpCandidates("8.2")
	if len(cands) != 2 || cands[0] != phpBinFor("8.2") {
		t.Errorf("phpCandidates = %v", cands)
	}
}

func TestPhpVersionParse(t *testing.T) {
	bin := fakeBin(t, "#!/bin/sh\necho 'PHP 8.2.31 (cli) (built: Feb 25 2025) Copyright (c) The PHP Group'\n")
	if got := phpVersion(bin); got != "8.2.31" {
		t.Errorf("phpVersion = %q, want 8.2.31", got)
	}
	if got := phpVersion(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Errorf("phpVersion missing = %q, want empty", got)
	}
	bad := fakeBin(t, "#!/bin/sh\necho 'not php'\n")
	if got := phpVersion(bad); got != "" {
		t.Errorf("phpVersion bad output = %q, want empty", got)
	}
}

func TestExtVersionAndLoaded(t *testing.T) {
	bin := fakeBin(t, "#!/bin/sh\necho 'swoole\nVersion => 5.1.5'\n")
	if got := extVersion(bin, "swoole"); got != "5.1.5" {
		t.Errorf("extVersion = %q, want 5.1.5", got)
	}
	if !extLoaded(bin, "swoole") {
		t.Error("extLoaded should be true")
	}
	if got := extVersion(filepath.Join(t.TempDir(), "missing"), "x"); got != "" {
		t.Errorf("extVersion missing = %q", got)
	}
}

func TestParseVer(t *testing.T) {
	v, ok := parseVer("8.2.31")
	if !ok || v.major != 8 || v.minor != 2 || v.patch != 31 {
		t.Errorf("parseVer 8.2.31 = %+v %v", v, ok)
	}
	v, ok = parseVer("8.2")
	if !ok || v.major != 8 || v.minor != 2 || v.patch != 0 {
		t.Errorf("parseVer 8.2 = %+v %v", v, ok)
	}
	if _, ok := parseVer("8"); ok {
		t.Error("parseVer single segment should fail")
	}
	if _, ok := parseVer(""); ok {
		t.Error("parseVer empty should fail")
	}
}

func TestCompareVer(t *testing.T) {
	a, _ := parseVer("8.2.31")
	b, _ := parseVer("8.3.0")
	c, _ := parseVer("8.2.5")
	if compareVer(a, b) >= 0 {
		t.Error("8.2.31 should be < 8.3.0")
	}
	if compareVer(b, a) <= 0 {
		t.Error("8.3.0 should be > 8.2.31")
	}
	if compareVer(a, c) <= 0 {
		t.Error("8.2.31 should be > 8.2.5")
	}
	if compareVer(a, a) != 0 {
		t.Error("equal versions should compare 0")
	}
}

func TestConstraintOk(t *testing.T) {
	cases := []struct {
		constraint, installed string
		want                  bool
	}{
		{"", "8.2.31", true},
		{"*", "8.2.31", true},
		{"8.2.31", "8.2.31", true},
		{"8.2", "8.2.0", true},
		{"8.3", "8.2.31", false},
		{">=8.2", "8.2.31", true},
		{">=8.3", "8.2.31", false},
		{"^8.2", "8.2.31", true},
		{"^8.3", "8.2.31", false},
		{"^9.0", "8.2.31", false},
		{"8.2", "garbage", false},
	}
	for _, c := range cases {
		if got := constraintOk(c.constraint, c.installed); got != c.want {
			t.Errorf("constraintOk(%q,%q) = %v, want %v", c.constraint, c.installed, got, c.want)
		}
	}
}

func TestServiceRunningNonexistent(t *testing.T) {
	if serviceRunning("__xpier_nonexistent__") {
		t.Error("serviceRunning should be false for a made-up service")
	}
}

func TestRunDispatch(t *testing.T) {
	homeTemp(t)
	// Pure commands must succeed without any environment.
	for _, cmd := range []string{"completion", "completion bash", "completion zsh", "help"} {
		if err := Run(strings.Split(cmd, " ")); err != nil {
			t.Errorf("Run(%q) = %v", cmd, err)
		}
	}
	if err := Run([]string{"no-such-command"}); err == nil {
		t.Error("unknown command should error")
	}
}

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}
