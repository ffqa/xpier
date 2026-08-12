package xpier

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

// cmdLog tails a site's php-fpm log (Herd's `herd log` equivalent).
func cmdLog(args []string) error {
	fs := flag.NewFlagSet("log", flag.ExitOnError)
	follow := fs.Bool("f", false, "follow the log")
	if err := fs.Parse(args); err != nil {
		return err
	}
	ver := ""
	if fs.NArg() > 0 {
		sites, err := store.LoadSites()
		if err != nil {
			return err
		}
		site, ok := sites.Sites[fs.Arg(0)]
		if !ok {
			return fmt.Errorf("site %s is not linked", fs.Arg(0))
		}
		ver = site.PHP
		if ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "php-fpm-"+ver+".log")
	if !store.FileExists(logPath) {
		return fmt.Errorf("log %s not found (start the site with `xpier sites:up`)", logPath)
	}
	if *follow {
		cmd := exec.Command("tail", "-f", logPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

// --- Mailpit (mail capture) ---

func mailpitBin() string {
	if p := filepath.Join(store.BrewPrefix(), "bin", "mailpit"); store.FileExists(p) {
		return p
	}
	return "/usr/local/bin/mailpit"
}

func mailStatePath() string { return filepath.Join(store.XpierHome(), "servers", "mailpit.json") }

func cmdMailUp(args []string) error {
	if st, err := loadMailState(); err == nil && store.PidAlive(st.PID) {
		fmt.Printf("mailpit already running (pid %d, UI http://127.0.0.1:8025)\n", st.PID)
		return nil
	}
	bin := mailpitBin()
	if err := store.EnsureBrewPackage(bin, "mailpit", "mailpit"); err != nil {
		return err
	}
	logPath := filepath.Join(store.XpierHome(), "logs", "mailpit.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logFile.Close()
	cmd := exec.Command(bin, "--smtp", "127.0.0.1:1025", "--listen", "127.0.0.1:8025")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = procAttr()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mailpit: %w", err)
	}
	writeMailState(cmd.Process.Pid, logPath)
	fmt.Printf("mailpit up (pid %d): SMTP 127.0.0.1:1025, UI http://127.0.0.1:8025\n", cmd.Process.Pid)
	return nil
}

func cmdMailDown(args []string) error {
	st, err := loadMailState()
	if err != nil {
		return fmt.Errorf("mailpit not running")
	}
	if store.PidAlive(st.PID) {
		store.KillGroup(st.PID, 15)
	}
	os.Remove(mailStatePath())
	fmt.Println("mailpit stopped")
	return nil
}

func cmdMail(args []string) error {
	return store.RunOutErr("open", "http://127.0.0.1:8025")
}

// --- Xdebug toggle ---

func xdebugConfPath(ver string) string {
	return filepath.Join(store.BrewPrefix(), "etc", "php", ver, "conf.d", "xpier-xdebug.ini")
}

func cmdXdebug(args []string) error {
	fs := flag.NewFlagSet("xdebug", flag.ExitOnError)
	ver := fs.String("php", nginx.DefaultPhpVersion(), "php version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	conf := xdebugConfPath(*ver)
	switch {
	case fs.NArg() == 0 || fs.Arg(0) == "status":
		if store.FileExists(conf) {
			fmt.Printf("xdebug ON (php@%s)\n", *ver)
		} else {
			fmt.Printf("xdebug OFF (php@%s)\n", *ver)
		}
	case fs.Arg(0) == "on":
		ext := filepath.Join(store.BrewPrefix(), "opt", "php@"+*ver, "lib", "php", "extensions", "no-debug-non-zts-*", "xdebug.so")
		if matches, _ := filepath.Glob(ext); len(matches) == 0 {
			return fmt.Errorf("xdebug extension not installed for php@%s (run `xpier sync --apply` or brew install shivammathur/extensions/xdebug@%s)", *ver, *ver)
		}
		content := "zend_extension=\"xdebug.so\"\nxdebug.mode=debug\nxdebug.start_with_request=yes\n"
		if err := os.MkdirAll(filepath.Dir(conf), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(conf, []byte(content), 0o644); err != nil {
			return err
		}
		fmt.Printf("xdebug ON (php@%s); restart php-fpm to apply\n", *ver)
	case fs.Arg(0) == "off":
		if err := os.Remove(conf); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Printf("xdebug OFF (php@%s); restart php-fpm to apply\n", *ver)
	default:
		return fmt.Errorf("usage: xpier xdebug [status|on|off] [--php 8.2]")
	}
	return nil
}

type mailState struct {
	PID     int    `json:"pid"`
	LogPath string `json:"log_path"`
}

func loadMailState() (*mailState, error) {
	data, err := os.ReadFile(mailStatePath())
	if err != nil {
		return nil, err
	}
	var st mailState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func writeMailState(pid int, logPath string) error {
	data, _ := json.Marshal(&mailState{PID: pid, LogPath: logPath})
	if err := os.MkdirAll(filepath.Dir(mailStatePath()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(mailStatePath(), data, 0o644)
}

func procAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
