package xpier

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

func xpierHome() string {
	// When run via `sudo xpier install`, keep using the real user's home.
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		if u, err := user.Lookup(sudoUser); err == nil {
			return filepath.Join(u.HomeDir, ".xpier")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return filepath.Join(home, ".xpier")
}

func slugFor(dir string) string {
	abs, _ := filepath.Abs(dir)
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:4])
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func killGroup(pid int, sig syscall.Signal) error {
	err := syscall.Kill(-pid, sig)
	if err == syscall.ESRCH {
		return syscall.Kill(pid, sig)
	}
	return err
}
