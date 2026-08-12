package xpier

import (
	"os"
	"path/filepath"
	"regexp"
)

var (
	serverNameRe = regexp.MustCompile(`'name'\s*=>\s*'([a-z0-9_]+)'`)
	serverPortRe = regexp.MustCompile(`'port'\s*=>\s*(\d+)`)
)

// serverPorts best-effort reads config/autoload/server.php for name -> port pairs.
func serverPorts(dir string) map[string]string {
	data, err := os.ReadFile(filepath.Join(dir, "config", "autoload", "server.php"))
	if err != nil {
		return nil
	}
	names := serverNameRe.FindAllStringSubmatch(string(data), -1)
	ports := serverPortRe.FindAllStringSubmatch(string(data), -1)
	if len(names) == 0 || len(ports) == 0 {
		return nil
	}
	out := make(map[string]string)
	for i := 0; i < len(names) && i < len(ports); i++ {
		out[names[i][1]] = ports[i][1]
	}
	return out
}
