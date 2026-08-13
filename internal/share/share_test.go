package share

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestShareStateRoundTrip(t *testing.T) {
	homeTemp(t)
	key := "larablog"
	if _, err := LoadShareState(key); err == nil {
		t.Error("LoadShareState missing should error")
	}
	st := &ShareState{Site: key, PID: 123, URL: "https://abc.trycloudflare.com", Target: "http://127.0.0.1:80", Log: "/tmp/l"}
	if err := SaveShareState(st); err != nil {
		t.Fatal(err)
	}
	got, err := LoadShareState(key)
	if err != nil {
		t.Fatal(err)
	}
	if got.PID != 123 || got.URL != "https://abc.trycloudflare.com" {
		t.Errorf("round trip = %+v", got)
	}
	if !strings.HasSuffix(ShareStatePath(key), "share-larablog.json") {
		t.Errorf("ShareStatePath = %q", ShareStatePath(key))
	}
}

func TestCloudflaredBin(t *testing.T) {
	homeTemp(t)
	if !strings.Contains(CloudflaredBin(), "cloudflared") {
		t.Errorf("CloudflaredBin = %q", CloudflaredBin())
	}
}

func TestTrycloudflareRe(t *testing.T) {
	log := "Registered tunnel connection conn=0 id=1 url=https://abc-123.trycloudflare.com"
	if got := trycloudflareRe.FindString(log); got != "https://abc-123.trycloudflare.com" {
		t.Errorf("regex match = %q", got)
	}
	if trycloudflareRe.FindString("no tunnel here") != "" {
		t.Error("regex should not match plain text")
	}
}

func TestProbeURLAndDetectOriginProto(t *testing.T) {
	homeTemp(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	_, portStr, err := splitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if code := probeURL("http://127.0.0.1:" + portStr); code != "200" {
		t.Errorf("probeURL = %q, want 200", code)
	}
	if proto := detectOriginProto(portStr); proto != "http" {
		t.Errorf("detectOriginProto = %q, want http", proto)
	}
	if code := probeURL("http://127.0.0.1:1"); code != "" {
		t.Errorf("probeURL closed port = %q, want empty", code)
	}
}

func TestCmdSharesEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdShares(nil); err != nil {
		t.Errorf("CmdShares = %v", err)
	}
}

func TestCmdShareStopEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdShareStop(nil); err == nil {
		t.Error("CmdShareStop with no shares should error")
	}
}

func TestCmdShareStaleStateStop(t *testing.T) {
	homeTemp(t)
	// A stale state file with a dead pid: stop should clean it up.
	st := &ShareState{Site: "ghost", PID: 999999, URL: "https://ghost.trycloudflare.com", Target: "http://127.0.0.1:80"}
	if err := SaveShareState(st); err != nil {
		t.Fatal(err)
	}
	if err := CmdShareStop([]string{"ghost"}); err != nil {
		t.Fatalf("CmdShareStop stale = %v", err)
	}
	if store.FileExists(ShareStatePath("ghost")) {
		t.Error("stale state not removed")
	}
}

func TestCmdShareUnknownBackend(t *testing.T) {
	homeTemp(t)
	if err := CmdShare([]string{"--backend", "ngrok", "xyz"}); err == nil {
		t.Error("unknown backend should error")
	}
}

func TestCmdShareUnlinkedSite(t *testing.T) {
	homeTemp(t)
	if err := CmdShare([]string{"nope"}); err == nil {
		t.Error("unlinked site should error")
	}
}

func splitHostPort(addr string) (string, string, error) {
	i := strings.LastIndexByte(addr, ':')
	if i < 0 {
		return "", "", os.ErrInvalid
	}
	if _, err := strconv.Atoi(addr[i+1:]); err != nil {
		return "", "", err
	}
	return addr[:i], addr[i+1:], nil
}

func TestWaitTunnelRegistered(t *testing.T) {
	homeTemp(t)
	log := filepath.Join(t.TempDir(), "tunnel.log")
	os.WriteFile(log, []byte("Registered tunnel connection conn=0"), 0o644)
	if !waitTunnelRegistered(log, 2*time.Second) {
		t.Error("waitTunnelRegistered should find the marker")
	}
	os.WriteFile(log, []byte("still connecting"), 0o644)
	if waitTunnelRegistered(log, 150*time.Millisecond) {
		t.Error("waitTunnelRegistered should time out without the marker")
	}
}

func TestStopShareByKeyStale(t *testing.T) {
	homeTemp(t)
	st := &ShareState{Site: "ghost", PID: 999999}
	SaveShareState(st)
	stopShareByKey("ghost")
	if store.FileExists(ShareStatePath("ghost")) {
		t.Error("stale state should be removed")
	}
	stopShareByKey("never-was")
}

func TestVerifyPublicURL(t *testing.T) {
	homeTemp(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	_, port, _ := splitHostPort(srv.Listener.Addr().String())
	if code := verifyPublicURL("http://127.0.0.1:" + port); code != "201" {
		t.Errorf("verifyPublicURL = %q, want 201", code)
	}
}

func TestCmdShareAlreadySharing(t *testing.T) {
	homeTemp(t)
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available for marker test")
	}
	s := store.DefaultSites()
	s.Sites["abc"] = store.Site{Path: "/srv/abc", Driver: "laravel"}
	s.Save()
	// A child whose cmdline contains the tunnel marker counts as sharing.
	child := exec.Command("python3", "-c", "import time;time.sleep(300)", "--url", "http://abc.test")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		child.Process.Kill()
		child.Wait()
	}()
	st := &ShareState{Site: "abc", PID: child.Process.Pid, URL: "https://abc.trycloudflare.com", Target: "http://abc.test"}
	SaveShareState(st)
	if err := CmdShare([]string{"abc"}); err != nil {
		t.Errorf("CmdShare already sharing = %v", err)
	}
}

func TestCmdShareStopLiveButNotOurs(t *testing.T) {
	homeTemp(t)
	st := &ShareState{Site: "ghost", PID: os.Getpid(), URL: "https://ghost.trycloudflare.com", Target: "http://127.0.0.1:80"}
	if err := SaveShareState(st); err != nil {
		t.Fatal(err)
	}
	if err := CmdShareStop([]string{"ghost"}); err != nil {
		t.Fatalf("CmdShareStop = %v", err)
	}
	if !store.PidAlive(os.Getpid()) {
		t.Fatal("CmdShareStop killed the test process!")
	}
	if store.FileExists(ShareStatePath("ghost")) {
		t.Error("stale share state should be removed")
	}
}

func TestAliveMarker(t *testing.T) {
	homeTemp(t)
	cf := &ShareState{Target: "http://127.0.0.1:80"}
	if got := aliveMarker(cf); got != "--url http://127.0.0.1:80" {
		t.Errorf("cloudflared marker = %q", got)
	}
	ssh := &ShareState{Kind: "localhost-run"}
	if got := aliveMarker(ssh); got != "nokey@localhost.run" {
		t.Errorf("ssh marker = %q", got)
	}
}

func TestSSHForwardSpec(t *testing.T) {
	homeTemp(t)
	cases := []struct{ target, sub, want string }{
		{"http://127.0.0.1:3000", "oauth", "oauth:80:localhost:3000"},
		{"http://127.0.0.1:3000", "", ":80:localhost:3000"},
		{"https://127.0.0.1:5173", "oauth", "oauth:443:localhost:5173"},
	}
	for _, c := range cases {
		if got := sshForwardSpec(c.target, c.sub); got != c.want {
			t.Errorf("sshForwardSpec(%q,%q) = %q, want %q", c.target, c.sub, got, c.want)
		}
	}
}

func TestServeoDedicatedKey(t *testing.T) {
	homeTemp(t)
	// No dedicated key: falls back to base URL.
	link := serveoRegisterURL()
	if link == "" || !strings.Contains(link, "console.serveo.net") {
		t.Errorf("fallback link = %q", link)
	}
	// Generate the dedicated key in the temp HOME and check the fingerprint.
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen not available")
	}
	dir := store.XpierHome() // temp HOME
	sshDir := filepath.Join(dir, "..", ".ssh")
	_ = sshDir
	home, _ := os.UserHomeDir()
	os.MkdirAll(filepath.Join(home, ".ssh"), 0o700)
	if out, err := exec.Command("ssh-keygen", "-t", "ed25519", "-f", filepath.Join(home, ".ssh", "id_serveo"), "-N", "", "-q").CombinedOutput(); err != nil {
		t.Skipf("ssh-keygen failed: %v %s", err, out)
	}
	priv, pub := serveoKeyFiles()
	if priv == "" || !strings.HasSuffix(pub, "id_serveo.pub") {
		t.Fatalf("serveoKeyFiles = %q %q", priv, pub)
	}
	link = serveoRegisterURL()
	if !strings.Contains(link, "add=SHA256%3A") {
		t.Errorf("dedicated link = %q", link)
	}
}
