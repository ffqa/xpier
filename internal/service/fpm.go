package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func FpmStatePath(ver string) string {
	return filepath.Join(store.XpierHome(), "servers", "fpm-"+ver+".json")
}

func FpmSockPath(ver string) string {
	return filepath.Join(store.XpierHome(), "run", "php-fpm-"+ver+".sock")
}

func FpmConfPath(ver string) string {
	return filepath.Join(store.XpierHome(), "fpm", "php-fpm-"+ver+".conf")
}

func FpmBinFor(ver string) string {
	return filepath.Join(store.BrewPrefix(), "opt", "php@"+ver, "sbin", "php-fpm")
}

func LoadFpmState(ver string) (*FpmState, error) {
	data, err := os.ReadFile(FpmStatePath(ver))
	if err != nil {
		return nil, err
	}
	var st FpmState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func FpmRunning(ver string) bool {
	if st, err := LoadFpmState(ver); err == nil {
		return store.PidAlive(st.PID)
	}
	return false
}

func WriteFpmConf(ver string) error {
	user, err := CurrentUser()
	if err != nil {
		return err
	}
	sock := FpmSockPath(ver)
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
	if err := os.MkdirAll(filepath.Dir(FpmConfPath(ver)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(FpmConfPath(ver), []byte(conf), 0o644)
}

// FpmUp starts (or reuses) a php-fpm child for a PHP version.
func FpmUp(ver string) error {
	if FpmRunning(ver) {
		return nil
	}
	bin := FpmBinFor(ver)
	if !store.FileExists(bin) {
		return fmt.Errorf("php-fpm for %s not found at %s (run `brew install shivammathur/php/php@%s`)", ver, bin, ver)
	}
	if err := WriteFpmConf(ver); err != nil {
		return err
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "php-fpm-"+ver+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	if err := os.MkdirAll(filepath.Dir(FpmSockPath(ver)), 0o755); err != nil {
		return err
	}
	cmd := exec.Command(bin, "-y", FpmConfPath(ver), "-F")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start php-fpm %s: %w", ver, err)
	}
	st := &FpmState{Version: ver, PID: cmd.Process.Pid, LogPath: logPath, Sock: FpmSockPath(ver)}
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(FpmStatePath(ver)), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(FpmStatePath(ver), data, 0o644); err != nil {
		return err
	}
	// Wait for the unix socket to appear.
	for i := 0; i < 20; i++ {
		if store.FileExists(FpmSockPath(ver)) {
			fmt.Printf("php-fpm %s up (pid %d, %s)\n", ver, cmd.Process.Pid, FpmSockPath(ver))
			return nil
		}
		if !store.PidAlive(cmd.Process.Pid) {
			return fmt.Errorf("php-fpm %s exited during startup; see %s", ver, logPath)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("php-fpm %s socket did not appear; see %s", ver, logPath)
}

func FpmDown(ver string) error {
	st, err := LoadFpmState(ver)
	if err != nil {
		return fmt.Errorf("php-fpm %s not running", ver)
	}
	if !store.PidAlive(st.PID) {
		os.Remove(FpmStatePath(ver))
		return nil
	}
	store.KillGroup(st.PID, syscall.SIGTERM)
	for i := 0; i < 50; i++ {
		if !store.PidAlive(st.PID) {
			os.Remove(FpmStatePath(ver))
			fmt.Printf("php-fpm %s stopped\n", ver)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	store.KillGroup(st.PID, syscall.SIGKILL)
	os.Remove(FpmStatePath(ver))
	fmt.Printf("php-fpm %s stopped (forced)\n", ver)
	return nil
}
