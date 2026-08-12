package apps

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"xpier/internal/nginx"
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
	if store.PidAlive(s.PID) {
		return true
	}
	return anyAppPortBusy(appPorts(app, s))
}

func strayAppPids(cmd string) []int {
	if cmd == "" {
		return nil
	}
	out, err := exec.Command("pgrep", "-f", cmd).Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if pid, err := strconv.Atoi(line); err == nil {
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

func killAppPortHolders(ports []string) {
	for _, p := range ports {
		out, _ := exec.Command("lsof", "-ti", "tcp:"+p, "-sTCP:LISTEN").Output()
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if pidStr := strings.TrimSpace(line); pidStr != "" {
				if pid, err := strconv.Atoi(pidStr); err == nil {
					syscall.Kill(pid, syscall.SIGKILL)
				}
			}
		}
	}
}

func appNginxConfPath(ns, name string) string {
	return filepath.Join(nginx.NginxConfDir(), "dev-"+ns+"-"+name+".conf")
}

func writeAppNginxConf(ns, name string, app store.App) error {
	if app.Domain == "" || app.Port == "" {
		return nil
	}
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
`, app.Domain, filepath.Join(store.XpierHome(), "certs", "wildcard.test.pem"),
		filepath.Join(store.XpierHome(), "certs", "wildcard.test-key.pem"), app.Port)
	return os.WriteFile(appNginxConfPath(ns, name), []byte(conf), 0o644)
}

func removeAppNginxConf(ns, name string) {
	os.Remove(appNginxConfPath(ns, name))
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
	s := &store.AppState{Name: name, PID: cmd.Process.Pid, Log: logPath, Port: app.Port, Ports: known, Domain: app.Domain}
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
	store.KillGroup(s.PID, syscall.SIGTERM)
	all := appPorts(app, s)
	for i := 0; i < 50; i++ {
		if !store.PidAlive(s.PID) && !anyAppPortBusy(all) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	killAppPortHolders(all)
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
	if out, err := exec.Command("brew", "install", "shivammathur/php/php@"+ver).CombinedOutput(); err != nil {
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
		if out, err := exec.Command("brew", "install", "shivammathur/extensions/"+ext+"@"+ver).CombinedOutput(); err != nil {
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

func CmdUp(args []string) error {
	cfg, cwd, err := LoadAppConfig()
	if err != nil {
		return err
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
		if len(strayAppPids(app.Cmd)) > 0 {
			conflicts = append(conflicts, fmt.Sprintf("%s 存在游离进程", n))
		}
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("namespace %q 已有进程在跑，拒绝重复启动：\n  %s\n请先 `xpier down`，或 `xpier restart <app>` 单独重启",
			ns, strings.Join(conflicts, "\n  "))
	}
	for n, app := range cfg.Apps {
		if err := appUp(ns, n, app); err != nil {
			fmt.Printf("  [warn] %v\n", err)
		}
	}
	if appConfigHasDomain(cfg) {
		for n, app := range cfg.Apps {
			writeAppNginxConf(ns, n, app)
		}
		nginx.NginxReload()
	}
	fmt.Printf("stack up (namespace %s). log: `xpier log <app>` | restart: `xpier restart <app>` | stop: `xpier down`\n", ns)
	_ = cwd
	return nil
}

func CmdDown(args []string) error {
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
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
			killAppPortHolders(cfgPorts)
			removeAppNginxConf(ns, n)
			fmt.Printf("  %s 端口被占但无状态（孤儿进程），已清理\n", n)
			any = true
			continue
		}
		if pids := strayAppPids(app.Cmd); len(pids) > 0 {
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
		} else if len(strayAppPids(app.Cmd)) > 0 {
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
				switch {
				case strings.HasPrefix(c, "up"):
					val = "\x1b[32m" + c + "\x1b[0m"
				case strings.HasPrefix(c, "down"):
					val = "\x1b[31m" + c + "\x1b[0m"
				default:
					val = "\x1b[33m" + c + "\x1b[0m"
				}
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
		return err
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
	writeAppNginxConf(ns, name, app)
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

func CmdLog(args []string) error {
	follow := false
	rest := make([]string, 0, len(args))
	for _, a := range args {
		if a == "-f" || a == "--follow" {
			follow = true
		} else {
			rest = append(rest, a)
		}
	}
	if len(rest) < 1 {
		return fmt.Errorf("usage: xpier log <app> [-f]")
	}
	cfg, _, err := LoadAppConfig()
	if err != nil {
		return err
	}
	ns := cfg.Namespace
	name := rest[0]
	s, err := store.LoadAppState(ns, name)
	if err != nil {
		return fmt.Errorf("app %s not running (start with `xpier up`)", name)
	}
	if follow {
		cmd := exec.Command("tail", "-f", s.Log)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	data, err := os.ReadFile(s.Log)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

var appLogColors = []string{"\x1b[36m", "\x1b[32m", "\x1b[33m", "\x1b[35m", "\x1b[34m"}

func CmdLogsAll(args []string) error {
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
	for i, n := range names {
		s, err := store.LoadAppState(ns, n)
		if err != nil {
			continue
		}
		color := appLogColors[i%len(appLogColors)]
		cmd := exec.Command("tail", "-f", s.Log)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}
		if err := cmd.Start(); err != nil {
			continue
		}
		defer cmd.Process.Kill()
		go func(name string, out io.Reader) {
			sc := bufio.NewScanner(out)
			sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for sc.Scan() {
				lineCh <- fmt.Sprintf("%s[%s]\x1b[0m %s", color, name, sc.Text())
			}
		}(n, stdout)
	}
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
