package apps

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"xpier/internal/nginx"
	"xpier/internal/service"
	"xpier/internal/sites"
	"xpier/internal/store"
)

// parseForceFlag extracts a --force flag from anywhere in args (Go's flag
// package stops at the first positional arg, silently ignoring a trailing
// --force - unsafe).
func parseForceFlag(args []string) (bool, []string) {
	force := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--force" {
			force = true
		} else {
			rest = append(rest, a)
		}
	}
	return force, rest
}

// store.App orchestration (merged from devstack): manage multiple dev servers
// (e.g. php-server/h5/admin) together, fully non-invasive to project code.

// LoadAppConfig reads apps from dev.yaml (devstack compat, namespace-aware)
// falling back to xpier.yaml's apps section.
func LoadAppConfig() (*store.AppConfig, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	devPath := filepath.Join(cwd, "dev.yaml")
	if store.FileExists(devPath) {
		data, err := os.ReadFile(devPath)
		if err != nil {
			return nil, "", err
		}
		var c store.AppConfig
		if err := store.YAMLUnmarshal(data, &c); err != nil {
			return nil, "", err
		}
		if len(c.Apps) == 0 {
			return nil, "", fmt.Errorf("no apps defined in %s", devPath)
		}
		if c.Namespace == "" {
			c.Namespace = "default"
		}
		return &c, cwd, nil
	}
	manifestPath, _ := store.ResolvePaths(cwd)
	m, err := store.LoadManifest(manifestPath)
	if err != nil || len(m.Apps) == 0 {
		return nil, "", fmt.Errorf("no apps defined (create dev.yaml or xpier.yaml with an apps: section)")
	}
	return &store.AppConfig{Namespace: "default", Apps: m.Apps}, cwd, nil
}

func appPortBusy(port string) bool {
	b, _ := store.PortBusy(port)
	return b
}

func anyAppPortBusy(ports []string) bool {
	for _, p := range ports {
		if appPortBusy(p) {
			return true
		}
	}
	return false
}

func appPorts(app store.App, s *store.AppState) []string {
	if len(app.Ports) > 0 {
		return app.Ports
	}
	if app.Port != "" {
		return []string{app.Port}
	}
	if len(s.Ports) > 0 {
		return s.Ports
	}
	if s.Port != "" {
		return []string{s.Port}
	}
	return nil
}

var appPortRe = regexp.MustCompile(`(?:localhost|127\.0\.0\.1|0\.0\.0\.0|\[::\]):(\d+)`)

func detectAppPorts(logPath string, known []string) []string {
	data, _ := os.ReadFile(logPath)
	var found []string
	seen := map[string]bool{}
	for _, m := range appPortRe.FindAllStringSubmatch(string(data), -1) {
		p := m[1]
		if p != "80" && p != "443" && !seen[p] {
			seen[p] = true
			found = append(found, p)
		}
	}
	if len(found) == 0 && len(known) > 0 {
		return known
	}
	return found
}

func appRunning(ns string, name string, app store.App) bool {
	s, err := store.LoadAppState(ns, name)
	if err != nil {
		return false
	}
	if store.ProcAlive(s.PID, s.Cmd) {
		return true
	}
	return anyAppPortBusy(appPorts(app, s))
}

// strayAppPids finds processes matching app.Cmd that are NOT tracked by any
// app state (a tracked pid belongs to a running app, not a stray) and not the
// xpier process itself.
func strayAppPids(ns, cmd string) []int {
	if cmd == "" {
		return nil
	}
	out, err := exec.Command("pgrep", "-f", cmd).Output()
	if err != nil {
		return nil
	}
	tracked := map[int]bool{}
	if entries, err := os.ReadDir(filepath.Join(store.XpierHome(), "apps", ns)); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".json") {
				if st, err := store.LoadAppState(ns, strings.TrimSuffix(e.Name(), ".json")); err == nil {
					tracked[st.PID] = true
				}
			}
		}
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if pid, err := strconv.Atoi(line); err == nil && pid != os.Getpid() && !tracked[pid] {
			pids = append(pids, pid)
		}
	}
	return pids
}

func killAppPids(pids []int) {
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGTERM)
	}
	time.Sleep(300 * time.Millisecond)
	for _, pid := range pids {
		syscall.Kill(pid, syscall.SIGKILL)
	}
}

// portHolderPids returns the PIDs listening on the given TCP ports.
func portHolderPids(ports []string) []int {
	var pids []int
	for _, p := range ports {
		out, _ := exec.Command("lsof", "-ti", "tcp:"+p, "-sTCP:LISTEN").Output()
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if pidStr := strings.TrimSpace(line); pidStr != "" {
				if pid, err := strconv.Atoi(pidStr); err == nil {
					pids = append(pids, pid)
				}
			}
		}
	}
	return pids
}

func procGroupOf(pid int) int {
	out, err := store.RunOut("ps", "-o", "pgid=", "-p", strconv.Itoa(pid))
	if err != nil {
		return 0
	}
	pgid, _ := strconv.Atoi(strings.TrimSpace(out))
	return pgid
}

// killAppPortHolders SIGKILLs port holders only when they belong to the app's
// process group (pgid from Setpgid). A holder in another group is never ours
// to kill, even if it occupies a configured port.
func killAppPortHolders(ports []string, pgid int) {
	if pgid <= 0 {
		return
	}
	for _, pid := range portHolderPids(ports) {
		if procGroupOf(pid) == pgid {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}

func appNginxConfPath(ns, name string) string {
	return filepath.Join(nginx.NginxConfDir(), "dev-"+ns+"-"+name+".conf")
}

func writeAppNginxConf(ns, name string, app store.App) error {
	port := app.Port
	if port == "" && len(app.Ports) > 0 {
		port = app.Ports[0]
	}
	// A running app may have detected a different actual port from its log
	// (e.g. vite re-binds when 5173 is taken); proxy to what is really serving.
	if st, err := store.LoadAppState(ns, name); err == nil {
		if st.Port != "" {
			port = st.Port
		} else if len(st.Ports) > 0 {
			port = st.Ports[0]
		}
	}
	if app.Domain == "" {
		return nil
	}
	if port == "" {
		return fmt.Errorf("cannot proxy %s: no port declared and none detected from the log; add port/ports to the app", app.Domain)
	}
	cert, certKey := nginx.CertPaths(nginx.CurrentTLD())
	conf := fmt.Sprintf(`server {
    listen 80;
    listen 443 ssl;
    server_name %s;
    ssl_certificate     %s;
    ssl_certificate_key %s;
    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
`, app.Domain, cert, certKey, port)
	if err := os.MkdirAll(filepath.Dir(appNginxConfPath(ns, name)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(appNginxConfPath(ns, name), []byte(conf), 0o644)
}

func removeAppNginxConf(ns, name string) {
	os.Remove(appNginxConfPath(ns, name))
}

// MigrateStateLogs rewrites app state log paths that still point at an old
// home (e.g. ~/.herdy -> ~/.xpier) to the computed location.
func MigrateStateLogs() error {
	root := filepath.Join(store.XpierHome(), "apps")
	nsEntries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	for _, ns := range nsEntries {
		if !ns.IsDir() {
			continue
		}
		states, err := os.ReadDir(filepath.Join(root, ns.Name()))
		if err != nil {
			continue
		}
		for _, f := range states {
			if !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			name := strings.TrimSuffix(f.Name(), ".json")
			st, err := store.LoadAppState(ns.Name(), name)
			if err != nil {
				continue
			}
			want := store.AppLogPath(ns.Name(), name)
			if st.Log != want {
				st.Log = want
				if err := store.SaveAppState(st, ns.Name()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// RefreshNginxConfs regenerates app proxy configs from saved app states
// (used after the xpier home directory moved, so cert paths are current).
func RefreshNginxConfs() error {
	root := filepath.Join(store.XpierHome(), "apps")
	nsEntries, err := os.ReadDir(root)
	if err != nil {
		return nil // no apps provisioned yet
	}
	for _, ns := range nsEntries {
		if !ns.IsDir() {
			continue
		}
		states, err := os.ReadDir(filepath.Join(root, ns.Name()))
		if err != nil {
			continue
		}
		for _, f := range states {
			if !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			name := strings.TrimSuffix(f.Name(), ".json")
			st, err := store.LoadAppState(ns.Name(), name)
			if err != nil {
				continue
			}
			if err := writeAppNginxConf(ns.Name(), name, store.App{Domain: st.Domain, Port: st.Port, Ports: st.Ports}); err != nil {
				fmt.Printf("[warn] %s: %v\n", name, err)
			}
		}
	}
	return nil
}

func appURL(app store.App, s *store.AppState) string {
	if app.Domain != "" {
		return "http://" + app.Domain + "/"
	}
	port := app.Port
	if port == "" && len(app.Ports) > 0 {
		port = app.Ports[0]
	}
	if s != nil {
		if s.Port != "" {
			port = s.Port
		} else if len(s.Ports) > 0 {
			port = s.Ports[0]
		}
	}
	if port != "" {
		return "http://127.0.0.1:" + port + "/"
	}
	return "-"
}

func appUp(ns string, name string, app store.App) error {
	if appRunning(ns, name, app) {
		fmt.Printf("  %s already up\n", name)
		return nil
	}
	known := appPorts(app, &store.AppState{})
	if anyAppPortBusy(known) {
		return fmt.Errorf("%s port(s) %s already in use", name, strings.Join(known, ","))
	}
	prepend, err := ensureAppPrereqs(app)
	if err != nil {
		return fmt.Errorf("%s 前置检查失败: %w", name, err)
	}
	logPath := store.AppLogPath(ns, name)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	cmd := exec.Command("sh", "-c", app.Cmd)
	cmd.Dir = app.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if len(app.Env) > 0 || prepend != "" {
		env := os.Environ()
		if prepend != "" {
			env = append(env, "PATH="+prepend+":"+os.Getenv("PATH"))
		}
		for k, v := range app.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}
	s := &store.AppState{Name: name, PID: cmd.Process.Pid, Cmd: app.Cmd, Log: logPath, Port: app.Port, Ports: known, Domain: app.Domain}
	if err := store.SaveAppState(s, ns); err != nil {
		store.KillGroup(cmd.Process.Pid, syscall.SIGKILL)
		return err
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if !store.PidAlive(s.PID) {
			return fmt.Errorf("%s exited during startup; see %s", name, logPath)
		}
		ports := appPorts(app, s)
		if len(ports) > 0 && anyAppPortBusy(ports) {
			break
		}
		if detected := detectAppPorts(logPath, known); len(detected) > 0 && anyAppPortBusy(detected) {
			s.Ports = detected
			s.Port = detected[0]
			store.SaveAppState(s, ns)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if s.Port == "" && len(app.Ports) == 0 {
		if detected := detectAppPorts(logPath, known); len(detected) > 0 {
			s.Ports = detected
			s.Port = detected[0]
			store.SaveAppState(s, ns)
		}
	}
	fmt.Printf("  %s up (pid %d, %s)\n", name, s.PID, appURL(app, s))
	return nil
}

func appDown(ns string, name string, app store.App) {
	s, err := store.LoadAppState(ns, name)
	if err != nil {
		return
	}
	all := appPorts(app, s)
	// Guard with a cmdline marker: never kill a PID recycled after reboot.
	if store.ProcAlive(s.PID, s.Cmd) {
		store.KillGroup(s.PID, syscall.SIGTERM)
		for i := 0; i < 50; i++ {
			if !store.PidAlive(s.PID) && !anyAppPortBusy(all) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	killAppPortHolders(all, s.PID)
	os.Remove(store.AppStatePath(ns, name))
	removeAppNginxConf(ns, name)
	fmt.Printf("  %s stopped\n", name)
}

// --- prereqs: node / php / extensions (guided auto-install) ---

func appParseMajor(v string) int {
	digits := ""
	for _, r := range v {
		if r >= '0' && r <= '9' {
			digits += string(r)
		} else if digits != "" {
			break
		}
	}
	n, _ := strconv.Atoi(digits)
	return n
}

func appNodeSatisfies(req string) bool {
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return false
	}
	return appParseMajor(string(out)) >= appParseMajor(req)
}

func appNvmNodeBinDir(req string) (string, bool) {
	home, _ := os.UserHomeDir()
	bases := []string{
		filepath.Join(home, ".nvm", "versions", "node"),
		filepath.Join(store.BrewPrefix(), "opt", "nvm", "versions", "node"),
		filepath.Join(home, "Library", "Application Support", "Herd", "config", "nvm", "versions", "node"),
	}
	for _, base := range bases {
		matches, _ := filepath.Glob(filepath.Join(base, "v"+req+".*", "bin"))
		if len(matches) > 0 {
			return matches[len(matches)-1], true
		}
	}
	return "", false
}

func appEnsureNode(req string) (string, error) {
	if appNodeSatisfies(req) {
		return "", nil
	}
	if dir, ok := appNvmNodeBinDir(req); ok {
		return dir, nil
	}
	ok, err := store.ConfirmYesNo(fmt.Sprintf("需要 node %s，但未安装。是否安装 nvm（brew install nvm）并 nvm install %s？", req, req))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("node %s 未安装（提示：装好 nvm 后 xpier 会自动切换）", req)
	}
	if out, err := exec.Command("brew", "install", "nvm").CombinedOutput(); err != nil {
		return "", fmt.Errorf("brew install nvm: %v: %s", err, out)
	}
	out, err := exec.Command("bash", "-lc",
		fmt.Sprintf("export NVM_DIR=\"$(brew --prefix nvm)\"; [ -s \"$NVM_DIR/nvm.sh\" ] && . \"$NVM_DIR/nvm.sh\"; nvm install %s; echo nvm-done", req)).CombinedOutput()
	if err != nil && !strings.Contains(string(out), "nvm-done") {
		return "", fmt.Errorf("nvm install %s 失败: %s", req, out)
	}
	if dir, ok := appNvmNodeBinDir(req); ok {
		return dir, nil
	}
	return "", fmt.Errorf("nvm install %s 后仍未找到 node %s，请重试（安装输出：%s）", req, req, strings.TrimSpace(string(out)))
}

func appEnsurePHP(ver string) error {
	bin := filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "bin", "php")
	if store.FileExists(bin) {
		return nil
	}
	ok, err := store.ConfirmYesNo(fmt.Sprintf("php@%s 未安装（brew install shivammathur/php/php@%s），是否安装？", ver, ver))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("php@%s 未安装", ver)
	}
	if out, err := store.RunOutYes("brew", "install", "shivammathur/php/php@"+ver); err != nil {
		return fmt.Errorf("brew install php@%s: %v: %s", ver, err, out)
	}
	return nil
}

func appEnsureExtensions(ver string, exts []string) error {
	bin := filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "bin", "php")
	for _, ext := range exts {
		out, err := exec.Command(bin, "-m").Output()
		if err == nil && strings.Contains(string(out), ext) {
			continue
		}
		ok, err := store.ConfirmYesNo(fmt.Sprintf("php@%s 缺少扩展 %s（brew install shivammathur/extensions/%s@%s），是否安装？", ver, ext, ext, ver))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("扩展 %s 未安装", ext)
		}
		store.BrewTrustTap("shivammathur/extensions")
		if out, err := store.RunOutYes("brew", "install", "shivammathur/extensions/"+ext+"@"+ver); err != nil {
			return fmt.Errorf("brew install %s@%s: %v: %s", ext, ver, err, out)
		}
	}
	return nil
}

func ensureAppPrereqs(app store.App) (string, error) {
	var prepend string
	if app.Node != "" {
		dir, err := appEnsureNode(app.Node)
		if err != nil {
			return "", err
		}
		prepend = dir
	}
	if app.PHP != "" {
		if err := appEnsurePHP(app.PHP); err != nil {
			return "", err
		}
		if len(app.Extensions) > 0 {
			if err := appEnsureExtensions(app.PHP, app.Extensions); err != nil {
				return "", err
			}
		}
	}
	return prepend, nil
}

func appConfigHasDomain(cfg *store.AppConfig) bool {
	for _, app := range cfg.Apps {
		if app.Domain != "" {
			return true
		}
	}
	return false
}

// --- commands ---

// autoLinkApp registers a web (fpm/static) app as a site: entries with a
// domain but no cmd are served by nginx+php-fpm like `xpier link`.
func autoLinkApp(ns, name string, app store.App, cwd string) error {
	dir := app.Dir
	if dir == "" {
		dir = cwd
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s: dir %s not found", name, abs)
	}
	if app.Domain == "" {
		return fmt.Errorf("%s: domain-only entries need a domain: field", name)
	}
	if !store.SafeSiteNameRe.MatchString(app.Domain) {
		return fmt.Errorf("%s: invalid domain %q", name, app.Domain)
	}
	reg, err := store.LoadSites()
	if err != nil {
		return err
	}
	if existing, ok := reg.Sites[name]; ok && existing.Path != abs {
		return fmt.Errorf("site %s already linked to %s (unlink it first)", name, existing.Path)
	}
	httpOnly := false
	site := store.Site{Path: abs, Driver: sites.DetectDriver(abs), PHP: app.PHP, Domain: app.Domain}
	if app.Secure {
		site.Secure = nil // https
	} else {
		site.Secure = &httpOnly // default http for app-declared sites
	}
	reg.Sites[name] = site
	if err := reg.Save(); err != nil {
		return err
	}
	if err := nginx.WriteSiteNginxConfig(reg, name); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	if store.IsTTY(os.Stdout) {
		phpVer := app.PHP
		if phpVer == "" {
			phpVer = nginx.DefaultPhpVersion()
		}
		if err := service.FpmUp(phpVer); err != nil {
			fmt.Printf("[warn] php-fpm %s: %v\n", phpVer, err)
		}
	}
	fmt.Printf("  %s site up: http://%s/\n", name, app.Domain)
	return nil
}

// upGuidance explains how a project type is meant to run when no apps are
// defined, instead of a bare "no apps defined" error.
func upGuidance() error {
	cwd, _ := os.Getwd()
	manifestPath, _ := store.ResolvePaths(cwd)
	m, err := store.LoadManifest(manifestPath)
	rt := "fpm"
	if err == nil {
		rt = m.Runtime
		if rt == "" {
			rt = "fpm"
		}
	}
	switch rt {
	case "hyperf", "swoole", "frankenphp":
		return fmt.Errorf("运行时 %s 需要 apps: 定义启动命令才能 `xpier up`;运行 `xpier app:init` 生成模板", rt)
	default:
		return fmt.Errorf("项目类型 %s 无需 `xpier up`(未定义 apps):站点由 nginx+php-fpm 直接服务;cd 项目 && xpier link,域名 = <目录名>.test", rt)
	}
}

func CmdUp(args []string) error {
	cfg, cwd, err := LoadAppConfig()
	if err != nil {
		return upGuidance()
	}
	ns := cfg.Namespace
	var conflicts []string
	for n, app := range cfg.Apps {
		if appRunning(ns, n, app) {
			conflicts = append(conflicts, fmt.Sprintf("%s(状态运行中)", n))
		}
		for _, p := range appPorts(app, &store.AppState{}) {
			if appPortBusy(p) {
				conflicts = append(conflicts, fmt.Sprintf("%s 端口 %s 已被占用", n, p))
			}
		}
		if len(strayAppPids(ns, app.Cmd)) > 0 {
			conflicts = append(conflicts, fmt.Sprintf("%s 存在游离进程", n))
		}
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("namespace %q 已有进程在跑，拒绝重复启动：\n  %s\n请先 `xpier down`，或 `xpier restart <app>` 单独重启",
			ns, strings.Join(conflicts, "\n  "))
	}
	for n, app := range cfg.Apps {
		if app.Cmd == "" {
			// Web (fpm/static) entry: `up` links it like xpier link.
			if err := autoLinkApp(ns, n, app, cwd); err != nil {
				fmt.Printf("  [warn] %v\n", err)
			}
			continue
		}
		if err := appUp(ns, n, app); err != nil {
			fmt.Printf("  [warn] %v\n", err)
		}
	}
	if appConfigHasDomain(cfg) {
		var failed []string
		var confErrs []string
		for n, app := range cfg.Apps {
			if app.Cmd == "" {
				continue // web entries already wrote their site config
			}
			if err := writeAppNginxConf(ns, n, app); err != nil {
				failed = append(failed, n)
				confErrs = append(confErrs, fmt.Sprintf("%s: %v", n, err))
			}
		}
		if len(confErrs) > 0 {
			// Roll back the apps whose proxy cannot be built so a failed
			// `up` never leaves half-started processes behind.
			for _, n := range failed {
				if appRunning(ns, n, cfg.Apps[n]) {
					appDown(ns, n, cfg.Apps[n])
				}
			}
			return fmt.Errorf("部分应用域名代理无法生成:\n  %s", strings.Join(confErrs, "\n  "))
		}
		nginx.NginxReload()
	}
	webApps, procApps := 0, 0
	for _, app := range cfg.Apps {
		if app.Cmd == "" {
			webApps++
		} else {
			procApps++
		}
	}
	switch {
	case webApps > 0 && procApps == 0:
		fmt.Printf("stack up (namespace %s): %d 个网站已注册,直接访问域名即可(`xpier sites` 查看)\n", ns, webApps)
	case webApps > 0:
		fmt.Printf("stack up (namespace %s): %d 网站 + %d 进程. 进程日志: `xpier app:log <app>` | 重启: `xpier restart <app>` | 停止: `xpier down`\n", ns, webApps, procApps)
	default:
		fmt.Printf("stack up (namespace %s). log: `xpier app:log <app>` | restart: `xpier restart <app>` | stop: `xpier down`\n", ns)
	}
	_ = cwd
	return nil
}

func CmdDown(args []string) error {
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return upGuidance()
	}
	ns := cfg.Namespace
	any := false
	for n, app := range cfg.Apps {
		if appRunning(ns, n, app) {
			appDown(ns, n, app)
			any = true
			continue
		}
		cfgPorts := appPorts(app, &store.AppState{})
		if anyAppPortBusy(cfgPorts) {
			var ours []int
			for _, pid := range portHolderPids(cfgPorts) {
				if store.ProcAlive(pid, app.Cmd) {
					ours = append(ours, pid)
				}
			}
			if len(ours) > 0 {
				killAppPids(ours)
				removeAppNginxConf(ns, n)
				fmt.Printf("  %s 端口被占且为游离进程（pid %v），已清理\n", n, ours)
				any = true
			} else {
				fmt.Printf("  [warn] %s 端口 %s 被其它进程占用（非 xpier 启动），跳过\n", n, strings.Join(cfgPorts, ","))
			}
			continue
		}
		if pids := strayAppPids(ns, app.Cmd); len(pids) > 0 {
			killAppPids(pids)
			removeAppNginxConf(ns, n)
			fmt.Printf("  %s 存在游离进程（pid %v），已清理\n", n, pids)
			any = true
		}
	}
	if any {
		if appConfigHasDomain(cfg) {
			nginx.NginxReload()
		}
		fmt.Println("stack down")
		return nil
	}
	webApps := 0
	for _, app := range cfg.Apps {
		if app.Cmd == "" {
			webApps++
		}
	}
	if webApps > 0 {
		fmt.Printf("没有进程型应用在跑;%d 个网站型应用仍是站点状态(`xpier sites` 查看,`xpier unlink <名字>` 移除)\n", webApps)
	} else {
		fmt.Println("no apps running")
	}
	return nil
}

type appStatusRow struct{ cells []string }

func CmdStatus(args []string) error {
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	names := make([]string, 0, len(cfg.Apps))
	for n := range cfg.Apps {
		names = append(names, n)
	}
	sort.Strings(names)
	showNode, showPHP, showExt := false, false, false
	for _, app := range cfg.Apps {
		if app.Node != "" {
			showNode = true
		}
		if app.PHP != "" {
			showPHP = true
		}
		if len(app.Extensions) > 0 {
			showExt = true
		}
	}
	var rows []appStatusRow
	for _, n := range names {
		app := cfg.Apps[n]
		state := "down"
		pid := "0"
		port := app.Port
		untracked := false
		if appRunning(ns, n, app) {
			state = "up"
		} else if anyAppPortBusy(appPorts(app, &store.AppState{})) {
			state = "up"
			untracked = true
		} else if len(strayAppPids(ns, app.Cmd)) > 0 {
			state = "up"
			untracked = true
		}
		var st *store.AppState
		if s, err := store.LoadAppState(ns, n); err == nil {
			st = s
			pid = strconv.Itoa(s.PID)
			if port == "" {
				port = s.Port
				if len(s.Ports) > 0 {
					port = strings.Join(s.Ports, ",")
				}
			}
		}
		if port == "" && len(app.Ports) > 0 {
			port = strings.Join(app.Ports, ",")
		}
		if untracked {
			state += "*"
		}
		cells := []string{n, state}
		if showNode {
			cells = append(cells, orDash(app.Node))
		}
		if showPHP {
			cells = append(cells, orDash(app.PHP))
		}
		if showExt {
			cells = append(cells, orDash(strings.Join(app.Extensions, ",")))
		}
		cells = append(cells, port, appURL(app, st), pid)
		rows = append(rows, appStatusRow{cells})
	}
	headers := []string{"name", "state"}
	if showNode {
		headers = append(headers, "node")
	}
	if showPHP {
		headers = append(headers, "php")
	}
	if showExt {
		headers = append(headers, "ext")
	}
	headers = append(headers, "port", "url", "pid")
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, c := range r.cells {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	printRow := func(cells []string) {
		line := "  "
		for i, c := range cells {
			val := c
			if i == 1 && c != "state" {
				val = store.Paint(c)
			}
			line += val + strings.Repeat(" ", widths[i]-len(c)+2)
		}
		fmt.Println(strings.TrimRight(line, " "))
	}
	printRow(headers)
	fmt.Println("  " + strings.Repeat("-", sumInts(widths)+2*(len(headers)-1)+1))
	for _, r := range rows {
		printRow(r.cells)
	}
	for _, r := range rows {
		if strings.HasSuffix(r.cells[1], "*") {
			fmt.Println("  (* = 未跟踪：端口被占或游离进程)")
			break
		}
	}
	return nil
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func sumInts(a []int) int {
	t := 0
	for _, v := range a {
		t += v
	}
	return t
}

func CmdStart(args []string) error {
	force, rest := parseForceFlag(args)
	if len(rest) < 1 {
		return fmt.Errorf("usage: xpier start <app> [--force]")
	}
	name := rest[0]
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return upGuidance()
	}
	ns := cfg.Namespace
	app, ok := cfg.Apps[name]
	if !ok {
		return fmt.Errorf("app %s not defined", name)
	}
	if appRunning(ns, name, app) {
		if !force {
			return fmt.Errorf("%s 已在运行；如需重启用 `xpier restart %s` 或 `xpier start %s --force`", name, name, name)
		}
		appDown(ns, name, app)
		clearAppCaches(app, force)
	} else if force {
		clearAppCaches(app, force)
	}
	if err := appUp(ns, name, app); err != nil {
		return err
	}
	if err := writeAppNginxConf(ns, name, app); err != nil {
		return err
	}
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	return nil
}

func CmdRestart(args []string) error {
	force, rest := parseForceFlag(args)
	if len(rest) < 1 {
		return fmt.Errorf("usage: xpier restart <app> [--force]")
	}
	name := rest[0]
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	app, ok := cfg.Apps[name]
	if !ok {
		return fmt.Errorf("app %s not defined", name)
	}
	appDown(ns, name, app)
	clearAppCaches(app, force)
	if err := appUp(ns, name, app); err != nil {
		return err
	}
	writeAppNginxConf(ns, name, app)
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	return nil
}

func clearAppCaches(app store.App, force bool) {
	if !force {
		return
	}
	ok, err := store.ConfirmYesNo(fmt.Sprintf("将删除 %s 下的编译缓存（runtime/container、bootstrap/cache，非代码），确认？", app.Dir))
	if err != nil {
		fmt.Printf("  [warn] %v\n", err)
		return
	}
	if !ok {
		fmt.Println("  skipped cache clearing")
		return
	}
	os.RemoveAll(filepath.Join(app.Dir, "runtime", "container"))
	bc := filepath.Join(app.Dir, "bootstrap", "cache")
	if entries, err := os.ReadDir(bc); err == nil {
		for _, e := range entries {
			n := e.Name()
			if strings.HasPrefix(n, "routes-") || n == "config.php" || n == "services.php" || n == "packages.php" {
				os.Remove(filepath.Join(bc, n))
			}
		}
	}
	fmt.Println("  cleared compiled caches")
}

// serviceLogPath resolves the log file for a built-in service name
// (nginx, dnsmasq, php-fpm[-<ver>], mailpit), or "" for app names.
func serviceLogPath(name string) string {
	switch {
	case name == "nginx":
		return filepath.Join(nginx.NginxHome(), "error.log")
	case name == "dnsmasq":
		return filepath.Join(store.XpierHome(), "logs", "com.xpier.dnsmasq.err.log")
	case strings.HasPrefix(name, "php-fpm"):
		ver := strings.TrimPrefix(name, "php-fpm-")
		if ver == name || ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
		return filepath.Join(store.XpierHome(), "logs", "php-fpm-"+ver+".log")
	case name == "mailpit":
		return filepath.Join(store.XpierHome(), "logs", "mailpit.log")
	}
	return ""
}

func tailFile(path string, follow bool, prefix string) error {
	if !store.FileExists(path) {
		return fmt.Errorf("log not found at %s", path)
	}
	if follow {
		args := []string{"-f"}
		if prefix != "" {
			args = append(args, "-n", "0")
		}
		cmd := exec.Command("tail", append(args, path)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if prefix != "" {
		for _, line := range strings.Split(string(data), "\n") {
			fmt.Printf("%s %s\n", prefix, line)
		}
		return nil
	}
	fmt.Print(string(data))
	return nil
}

// parseLogFlags extracts -f/--follow, leaving the rest untouched.
func isHelpArg(args []string) bool {
	return len(args) > 0 && (args[0] == "-h" || args[0] == "--help")
}

func parseLogFlags(args []string) (follow bool, rest []string) {
	for _, a := range args {
		if a == "-f" || a == "--follow" {
			follow = true
		} else {
			rest = append(rest, a)
		}
	}
	return follow, rest
}

// CmdLog tails a single built-in service log (global, no project needed).
func CmdLog(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier log <nginx|dnsmasq|php-fpm|mailpit> [-f]")
		return nil
	}
	follow, rest := parseLogFlags(args)
	if len(rest) < 1 {
		return fmt.Errorf("usage: xpier log <nginx|dnsmasq|php-fpm|mailpit> [-f] (project app logs: `xpier app:log <app>`)")
	}
	name := rest[0]
	path := serviceLogPath(name)
	if path == "" {
		return fmt.Errorf("unknown service %q (nginx|dnsmasq|php-fpm|mailpit); for app logs use `xpier app:log %s`", name, name)
	}
	return tailFile(path, follow, "")
}

// CmdAppLog tails one running app's log in the current project directory.
func CmdAppLog(args []string) error {
	if isHelpArg(args) {
		fmt.Println("usage: xpier app:log <app> [-f]")
		return nil
	}
	follow, rest := parseLogFlags(args)
	if len(rest) < 1 {
		return fmt.Errorf("usage: xpier app:log <app> [-f]")
	}
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	name := rest[0]
	if app, ok := cfg.Apps[name]; ok && app.Cmd == "" {
		return fmt.Errorf("%s 是网站型应用(无进程日志);直接访问 %s 即可", name, appURL(app, nil))
	}
	s, err := store.LoadAppState(ns, name)
	if err != nil {
		return fmt.Errorf("app %s not running (start with `xpier up`)", name)
	}
	path := s.Log
	if path == "" || !store.FileExists(path) {
		path = store.AppLogPath(ns, name)
	}
	return tailFile(path, follow, "")
}

var appLogColors = []string{"\x1b[36m", "\x1b[32m", "\x1b[33m", "\x1b[35m", "\x1b[34m"}

// tailAllLogs tails the running apps plus nginx + php-fpm service logs
// (Herd's `herd logs` shows everything, nginx included).
type logEntry struct {
	path   string
	prefix string
}

func serviceLogEntries() []logEntry {
	var out []logEntry
	for _, svc := range []string{"nginx", "dnsmasq", "php-fpm", "mailpit"} {
		if p := serviceLogPath(svc); p != "" && store.FileExists(p) {
			out = append(out, logEntry{p, "[" + svc + "]"})
		}
	}
	return out
}

// tailEntries tails the given log files concurrently with colored prefixes.
func tailEntries(entries []logEntry) error {
	if len(entries) == 0 {
		return fmt.Errorf("no logs to tail (service logs may not exist yet)")
	}
	lineCh := make(chan string, 64)
	var wg sync.WaitGroup
	started := 0
	for _, e := range entries {
		cmd := exec.Command("tail", "-f", "-n", "0", e.path)
		cmd.Stderr = os.Stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}
		if err := cmd.Start(); err != nil {
			continue
		}
		started++
		defer cmd.Process.Kill()
		wg.Add(1)
		idx := started - 1
		go func(p, name string, out io.Reader) {
			defer wg.Done()
			sc := bufio.NewScanner(out)
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			color := appLogColors[idx%len(appLogColors)]
			for sc.Scan() {
				lineCh <- fmt.Sprintf("%s%s\x1b[0m %s", color, name, sc.Text())
			}
		}(e.path, e.prefix, stdout)
	}
	if started == 0 {
		return fmt.Errorf("no logs could be tailed")
	}
	go func() {
		wg.Wait()
		close(lineCh)
	}()
	for line := range lineCh {
		fmt.Println(line)
	}
	return nil
}

// tailServiceLogs tails the global service logs (nginx, dnsmasq, php-fpm,
// mailpit) - usable from any directory.
func tailServiceLogs() error {
	return tailEntries(serviceLogEntries())
}

// tailAllLogs tails service logs plus the current project's app logs.
func tailAllLogs() error {
	entries := serviceLogEntries()
	if cfg, _, err := LoadAppConfig(); err == nil {
		ns := cfg.Namespace
		for n, app := range cfg.Apps {
			if !appRunning(ns, n, app) {
				continue
			}
			path := store.AppLogPath(ns, n)
			if s, err := store.LoadAppState(ns, n); err == nil {
				path = s.Log
				if path == "" || !store.FileExists(path) {
					path = store.AppLogPath(ns, n)
				}
			}
			if store.FileExists(path) {
				entries = append(entries, logEntry{path, "[" + n + "]"})
			}
		}
	}
	return tailEntries(entries)
}

// CmdLogsAll tails service logs (global view, no project needed). With
// "all" it also includes the current project's running app logs.
func CmdLogsAll(args []string) error {
	if len(args) > 0 && args[0] == "all" {
		return tailAllLogs()
	}
	return tailServiceLogs()
}

// CmdAppLogsAll tails the current project's running apps (project-scoped).
func CmdAppLogsAll(args []string) error {
	return tailAppLogs()
}

func tailAppLogs() error {
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	names := make([]string, 0, len(cfg.Apps))
	for n, app := range cfg.Apps {
		if appRunning(ns, n, app) {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return fmt.Errorf("no apps running (start with `xpier up`)")
	}
	fmt.Printf("tailing %d app(s) - Ctrl-C to stop\n", len(names))
	lineCh := make(chan string, 64)
	var wg sync.WaitGroup
	started := 0
	for i, n := range names {
		s, err := store.LoadAppState(ns, n)
		if err != nil {
			continue
		}
		color := appLogColors[i%len(appLogColors)]
		logPath := s.Log
		if logPath == "" || !store.FileExists(logPath) {
			// State written before a home rename (e.g. ~/.herdy -> ~/.xpier)
			// can point at a moved log; fall back to the computed path.
			logPath = store.AppLogPath(ns, n)
		}
		if !store.FileExists(logPath) {
			fmt.Printf("[warn] %s: log not found at %s\n", n, logPath)
			continue
		}
		cmd := exec.Command("tail", "-f", logPath)
		cmd.Stderr = os.Stderr // surface tail errors instead of failing silently
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}
		if err := cmd.Start(); err != nil {
			continue
		}
		started++
		defer cmd.Process.Kill()
		wg.Add(1)
		go func(name string, out io.Reader) {
			defer wg.Done()
			sc := bufio.NewScanner(out)
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for sc.Scan() {
				lineCh <- fmt.Sprintf("%s[%s]\x1b[0m %s", color, name, sc.Text())
			}
		}(n, stdout)
	}
	if started == 0 {
		return fmt.Errorf("no app logs to tail")
	}
	go func() {
		wg.Wait()
		close(lineCh)
	}()
	for line := range lineCh {
		fmt.Println(line)
	}
	return nil
}

func CmdURL(args []string) error {
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	show := func(n string, app store.App) {
		var st *store.AppState
		if s, err := store.LoadAppState(ns, n); err == nil {
			st = s
		}
		fmt.Printf("  %-12s %s\n", n, appURL(app, st))
	}
	if len(args) > 0 {
		app, ok := cfg.Apps[args[0]]
		if !ok {
			return fmt.Errorf("app %s not defined", args[0])
		}
		show(args[0], app)
		return nil
	}
	for n, app := range cfg.Apps {
		show(n, app)
	}
	return nil
}

// appInitTemplate is the commented dev.yaml skeleton written by `app:init`.
const appInitTemplate = `# xpier 应用编排配置(dev.yaml)
# 放在项目根目录;xpier up / start / log / url 从这里读取应用定义。
# 字段说明见 docs/architecture.md,命令见 docs/commands.md。

namespace: devstack          # 可选:进程隔离命名空间(默认 default;不同项目用不同 namespace 可同时跑)

apps:
  # 示例 0:标准 Laravel / 静态站点 —— 无需 cmd 和端口!
  # up 会自动把它注册成站点(相当于 xpier link),nginx+php-fpm 直接服务
  blog:
    dir: /path/to/laravel-blog   # 可选:默认当前目录
    domain: blog.test            # 必填:访问域名
    secure: true                 # 可选:开启 https(默认 http)
    php: "8.4"                   # 可选:固定 PHP 版本

  # 示例 1:PHP 常驻服务(Hyperf/Swoole)
  php-server:
    dir: /path/to/php-server  # 必填:工作目录(命令在此执行)
    cmd: php bin/hyperf.php server:watch   # 必填:启动命令
    ports: ["9501", "9502"]   # 可选:监听的端口(状态检测、端口冲突判断)
    php: "8.2"                # 可选:固定 PHP 版本(启动前确保已安装)
    extensions: [swoole, redis]  # 可选:需要的 PHP 扩展(启动前检查)
    domain: api.test          # 可选:生成 nginx 反代,访问 http://api.test

  # 示例 2:前端应用
  h5:
    dir: /path/to/h5
    cmd: npm run dev:test
    node: "20"                # 可选:经 nvm 固定 Node 大版本
    env:                      # 可选:注入环境变量
      VITE_OPEN: "0"

  # 示例 3:单端口应用(不配 domain 就没有 nginx 映射,仅托管进程)
  admin:
    dir: /path/to/admin
    cmd: npm run dev:local -- --no-open
    port: "5173"
`

// appInitGuide is printed before writing the template so users who have never
// seen dev.yaml know what to edit.
const appInitGuide = `接下来会生成一个带注释的 dev.yaml 模板,照着下面改:

1. namespace —— 一般不用动;不同项目想同时运行且互不干扰时才改成不同名字。
2. apps 下的每个条目 = 一个应用。复制一份示例改成自己的:
   - dir    必填,指向应用所在目录
   - cmd    必填,平时你手动在终端敲的那条启动命令
   - ports  填它监听的端口(没有就不填)
   - php/node 可选,固定运行时版本
   - domain 可选,配了之后自动生成 nginx 反代,浏览器直接访问该域名
3. 改完运行 xpier up(全起)/ xpier start <app>(单个)/ xpier log <app>(看日志)。

`

// CmdInit generates a commented dev.yaml template for a group directory.
// --force overwrites an existing dev.yaml.
func CmdInit(args []string) error {
	fs := flag.NewFlagSet("app:init", flag.ExitOnError)
	force := fs.Bool("force", false, "overwrite an existing dev.yaml with the template")
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", abs)
	}
	target := filepath.Join(abs, "dev.yaml")
	if store.FileExists(target) && !*force {
		return fmt.Errorf("%s already exists (use --force to overwrite with the template)", target)
	}
	fmt.Print(appInitGuide)
	if err := os.WriteFile(target, []byte(appInitTemplate), 0o644); err != nil {
		return err
	}
	fmt.Printf("created %s\n", target)
	fmt.Println("edit it, then run `xpier up`")
	return nil
}
