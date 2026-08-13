package service

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"xpier/internal/nginx"
	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestFpmPaths(t *testing.T) {
	homeTemp(t)
	home := store.XpierHome()
	if FpmStatePath("8.2") != filepath.Join(home, "servers", "fpm-8.2.json") {
		t.Errorf("FpmStatePath = %q", FpmStatePath("8.2"))
	}
	if FpmSockPath("8.2") != filepath.Join(home, "run", "php-fpm-8.2.sock") {
		t.Errorf("FpmSockPath = %q", FpmSockPath("8.2"))
	}
	if FpmConfPath("8.2") != filepath.Join(home, "fpm", "php-fpm-8.2.conf") {
		t.Errorf("FpmConfPath = %q", FpmConfPath("8.2"))
	}
	if !strings.HasSuffix(FpmBinFor("8.2"), "php@8.2/sbin/php-fpm") {
		t.Errorf("FpmBinFor = %q", FpmBinFor("8.2"))
	}
}

func TestWriteFpmConf(t *testing.T) {
	homeTemp(t)
	if err := WriteFpmConf("8.2"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(FpmConfPath("8.2"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"[global]", "[www]", "daemonize = no", "pm.max_children = 8", "listen = " + FpmSockPath("8.2")} {
		if !strings.Contains(s, want) {
			t.Errorf("fpm conf missing %q:\n%s", want, s)
		}
	}
}

func TestFpmStateRoundTripAndRunning(t *testing.T) {
	homeTemp(t)
	ver := "8.2"
	if FpmRunning(ver) {
		t.Error("FpmRunning should be false with no state")
	}
	st := &FpmState{Version: ver, PID: os.Getpid(), LogPath: "/tmp/x.log", Sock: "/tmp/x.sock"}
	if err := os.MkdirAll(filepath.Dir(FpmStatePath(ver)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(st)
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFpmState(ver)
	if err != nil {
		t.Fatal(err)
	}
	if got.PID != os.Getpid() || got.Sock != "/tmp/x.sock" {
		t.Errorf("round trip = %+v", got)
	}
	// The test process is alive but its cmdline lacks the fpm marker: the
	// guard must NOT report it as a running php-fpm (recycled-PID protection).
	if FpmRunning(ver) {
		t.Error("FpmRunning should reject a live pid without the fpm marker")
	}
	// A real child whose cmdline contains the marker is recognized.
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available for marker test")
	}
	marker := FpmConfPath(ver)
	child := exec.Command("python3", "-c", "import time;time.sleep(300)", marker)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		child.Process.Kill()
		child.Wait()
	}()
	st.PID = child.Process.Pid
	data, err = json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if !FpmRunning(ver) {
		t.Error("FpmRunning should be true for a live pid with the fpm marker")
	}
	// Remove the marker child so the dead-pid case cannot match it.
	child.Process.Kill()
	child.Wait()
	// Dead pid -> false.
	st.PID = 999999
	data, err = json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if FpmRunning(ver) {
		t.Error("FpmRunning should be false for dead pid")
	}
}

func TestFpmDown(t *testing.T) {
	homeTemp(t)
	if err := FpmDown("8.2"); err == nil {
		t.Error("FpmDown without state should error")
	}
	// State with dead pid: removes file, no error.
	ver := "8.2"
	st := &FpmState{Version: ver, PID: 999999}
	if err := os.MkdirAll(filepath.Dir(FpmStatePath(ver)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(st)
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := FpmDown(ver); err != nil {
		t.Fatalf("FpmDown dead pid = %v", err)
	}
	if store.FileExists(FpmStatePath(ver)) {
		t.Error("state file should be removed")
	}
}

func TestLaunchdPlist(t *testing.T) {
	homeTemp(t)
	p := LaunchdPlist("com.xpier.test", "/bin/true", "-x")
	for _, want := range []string{"<string>com.xpier.test</string>", "<string>/bin/true</string>", "<string>-x</string>", "RunAtLoad", "KeepAlive"} {
		if !strings.Contains(p, want) {
			t.Errorf("plist missing %q:\n%s", want, p)
		}
	}
	np := LaunchdPlistNginx()
	if !strings.Contains(np, "com.xpier.nginx") {
		t.Error("nginx plist label missing")
	}
	dp := LaunchdPlistDnsmasq()
	if !strings.Contains(dp, "com.xpier.dnsmasq") {
		t.Error("dnsmasq plist label missing")
	}
}

func TestLaunchdDir(t *testing.T) {
	if LaunchdDir() != "/Library/LaunchDaemons" {
		t.Errorf("LaunchdDir = %q", LaunchdDir())
	}
}

func TestEnsureWildcardCert(t *testing.T) {
	homeTemp(t)
	if err := EnsureWildcardCert("test"); err != nil {
		t.Skipf("openssl unavailable: %v", err)
	}
	cert, key := nginx.CertPaths("test")
	if !store.FileExists(cert) || !store.FileExists(key) {
		t.Error("wildcard cert not created")
	}
	// Second call is a no-op.
	if err := EnsureWildcardCert("test"); err != nil {
		t.Errorf("second EnsureWildcardCert = %v", err)
	}
}

func TestDnsmasqBin(t *testing.T) {
	homeTemp(t)
	bin := DnsmasqBin()
	if bin == "" {
		t.Error("DnsmasqBin empty")
	}
	if !strings.Contains(bin, "dnsmasq") {
		t.Errorf("DnsmasqBin = %q", bin)
	}
}

func TestCmdServicesEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdServices(nil); err != nil {
		t.Errorf("CmdServices = %v", err)
	}
}

func TestCmdServiceUnknown(t *testing.T) {
	homeTemp(t)
	if err := CmdService([]string{"bogus", "status"}); err == nil {
		t.Error("unknown service should error")
	}
	if err := CmdService([]string{"nginx"}); err == nil {
		t.Error("missing action should error")
	}
	if err := CmdService([]string{"nginx", "bogus-action"}); err == nil {
		t.Error("unknown action should error")
	}
}

func TestCmdServiceStatuses(t *testing.T) {
	homeTemp(t)
	if err := CmdService([]string{"nginx", "status"}); err != nil {
		t.Errorf("nginx status = %v", err)
	}
	if err := CmdService([]string{"dnsmasq", "status"}); err != nil {
		t.Errorf("dnsmasq status = %v", err)
	}
	if err := CmdService([]string{"php-fpm", "status"}); err != nil {
		t.Errorf("php-fpm status = %v", err)
	}
}

func TestCmdServicesStopStartEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdServicesStop(nil); err != nil {
		t.Errorf("CmdServicesStop = %v", err)
	}
	if err := CmdServicesStart(nil); err != nil {
		t.Errorf("CmdServicesStart = %v", err)
	}
}

func TestCurrentUser(t *testing.T) {
	homeTemp(t)
	u, err := store.CurrentUser()
	if err != nil || u.Username == "" {
		t.Errorf("CurrentUser = %+v %v", u, err)
	}
}

func TestCmdServiceConfigMissing(t *testing.T) {
	homeTemp(t)
	if err := CmdService([]string{"nginx", "config"}); err == nil {
		t.Error("nginx config with missing conf should error")
	}
	if err := CmdService([]string{"dnsmasq", "config"}); err == nil {
		t.Error("dnsmasq config with missing conf should error")
	}
	if err := CmdService([]string{"php-fpm", "config"}); err == nil {
		t.Error("php-fpm config with missing conf should error")
	}
}

func TestCmdServicesStartWithLinkedSitesNoFpm(t *testing.T) {
	homeTemp(t)
	// A linked site whose php-fpm is not installed: warn, but no error.
	s := store.DefaultSites()
	s.Sites["abc"] = store.Site{Path: "/srv/abc", Driver: "laravel", PHP: "99.9"}
	s.Save()
	if err := CmdServicesStart(nil); err != nil {
		t.Errorf("CmdServicesStart warn path = %v", err)
	}
}

func TestFpmDownLiveButNotOurs(t *testing.T) {
	homeTemp(t)
	ver := "8.2"
	// A live PID (the test process) whose cmdline does not match the fpm
	// marker: FpmDown must NOT kill it, only drop the stale state.
	st := &FpmState{Version: ver, PID: os.Getpid()}
	if err := os.MkdirAll(filepath.Dir(FpmStatePath(ver)), 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(st)
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := FpmDown(ver); err != nil {
		t.Fatalf("FpmDown = %v", err)
	}
	if !store.PidAlive(os.Getpid()) {
		t.Fatal("FpmDown killed the test process!")
	}
	if store.FileExists(FpmStatePath(ver)) {
		t.Error("stale fpm state should be removed")
	}
}

func TestCmdPhpList(t *testing.T) {
	homeTemp(t)
	if err := CmdPhpList(nil); err != nil {
		t.Fatalf("CmdPhpList = %v", err)
	}
}
