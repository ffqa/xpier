package xpier

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"xpier/internal/store"
)

type FpmState struct {
	Version string `json:"version"`
	PID     int    `json:"pid"`
	LogPath string `json:"log_path"`
	Sock    string `json:"sock"`
}

func fpmStatePath(ver string) string {
	return filepath.Join(store.XpierHome(), "servers", "fpm-"+ver+".json")
}

func fpmSockPath(ver string) string {
	return filepath.Join(store.XpierHome(), "run", "php-fpm-"+ver+".sock")
}

func fpmConfPath(ver string) string {
	return filepath.Join(store.XpierHome(), "fpm", "php-fpm-"+ver+".conf")
}

func fpmBinFor(ver string) string {
	return filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "sbin", "php-fpm")
}

func loadFpmState(ver string) (*FpmState, error) {
	data, err := os.ReadFile(fpmStatePath(ver))
	if err != nil {
		return nil, err
	}
	var st FpmState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func fpmRunning(ver string) bool {
	if st, err := loadFpmState(ver); err == nil {
		return store.PidAlive(st.PID)
	}
	return false
}

func writeFpmConf(ver string) error {
	user, err := currentUser()
	if err != nil {
		return err
	}
	sock := fpmSockPath(ver)
	conf := fmt.Sprintf(`[global]
pid = %s/run/php-fpm-%s.pid
error_log = %s/logs/php-fpm-%s.log
daemonize = no

[www]
user = %s
group = %s
listen = %s
listen.owner = %s
listen.group = %s
listen.mode = 0660
pm = dynamic
pm.max_children = 8
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 4
clear_env = no
`, store.XpierHome(), ver, store.XpierHome(), ver, user.Username, user.Gid, sock, user.Username, user.Gid)
	if err := os.MkdirAll(filepath.Dir(fpmConfPath(ver)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fpmConfPath(ver), []byte(conf), 0o644)
}

// fpmUp starts (or reuses) a php-fpm child for a PHP version.
func fpmUp(ver string) error {
	if fpmRunning(ver) {
		return nil
	}
	bin := fpmBinFor(ver)
	if !store.FileExists(bin) {
		return fmt.Errorf("php-fpm for %s not found at %s (run `brew install shivammathur/php/php@%s`)", ver, bin, ver)
	}
	if err := writeFpmConf(ver); err != nil {
		return err
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "php-fpm-"+ver+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	if err := os.MkdirAll(filepath.Dir(fpmSockPath(ver)), 0o755); err != nil {
		return err
	}
	cmd := exec.Command(bin, "-y", fpmConfPath(ver), "-F")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start php-fpm %s: %w", ver, err)
	}
	st := &FpmState{Version: ver, PID: cmd.Process.Pid, LogPath: logPath, Sock: fpmSockPath(ver)}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fpmStatePath(ver)), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(fpmStatePath(ver), data, 0o644); err != nil {
		return err
	}
	// Wait for the unix socket to appear.
	for i := 0; i < 20; i++ {
		if store.FileExists(fpmSockPath(ver)) {
			fmt.Printf("php-fpm %s up (pid %d, %s)\n", ver, cmd.Process.Pid, fpmSockPath(ver))
			return nil
		}
		if !store.PidAlive(cmd.Process.Pid) {
			return fmt.Errorf("php-fpm %s exited during startup; see %s", ver, logPath)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("php-fpm %s socket did not appear; see %s", ver, logPath)
}

func fpmDown(ver string) error {
	st, err := loadFpmState(ver)
	if err != nil {
		return fmt.Errorf("php-fpm %s not running", ver)
	}
	if !store.PidAlive(st.PID) {
		os.Remove(fpmStatePath(ver))
		return nil
	}
	store.KillGroup(st.PID, syscall.SIGTERM)
	for i := 0; i < 50; i++ {
		if !store.PidAlive(st.PID) {
			os.Remove(fpmStatePath(ver))
			fmt.Printf("php-fpm %s stopped\n", ver)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	store.KillGroup(st.PID, syscall.SIGKILL)
	os.Remove(fpmStatePath(ver))
	fmt.Printf("php-fpm %s stopped (forced)\n", ver)
	return nil
}

func cmdSitesUp(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	syncParked(sites)
	if err := sites.Save(); err != nil {
		return err
	}
	if len(sites.Sites) == 0 {
		return fmt.Errorf("no linked sites; link one with `xpier link` or `xpier park <dir>` first")
	}
	if b, _ := store.PortBusy("80"); !b {
		fmt.Println("[warn] nginx is not listening on port 80; run `sudo xpier install` first")
	}
	versions := map[string]bool{}
	for _, site := range sites.Sites {
		ver := site.PHP
		if ver == "" {
			ver = defaultPhpVersion()
		}
		versions[ver] = true
	}
	for ver := range versions {
		if err := fpmUp(ver); err != nil {
			fmt.Printf("[warn] %v\n", err)
		}
	}
	return nil
}

func cmdSitesDown(args []string) error {
	entries, err := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	if err != nil {
		return fmt.Errorf("no php-fpm state found")
	}
	stopped := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			if err := fpmDown(ver); err == nil {
				stopped++
			}
		}
	}
	if stopped == 0 {
		fmt.Println("no php-fpm running")
	}
	return nil
}
