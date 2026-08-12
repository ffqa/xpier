package xpier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func dnsmasqRunning() bool {
	// lsof sometimes fails to report dnsmasq's UDP 53 socket on macOS;
	// detect the process instead.
	out, err := runOut("pgrep", "-f", dnsmasqBin())
	return err == nil && strings.TrimSpace(out) != ""
}

func xpierServiceStatus() {
	nginxUp := false
	if b, _ := portBusy("80"); b {
		nginxUp = true
	}
	dnsUp := dnsmasqRunning()
	fmt.Printf("nginx:   %s\n", upDown(nginxUp))
	fmt.Printf("dnsmasq: %s\n", upDown(dnsUp))
	entries, _ := os.ReadDir(filepath.Join(xpierHome(), "servers"))
	fpm := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			if fpmRunning(ver) {
				fpm = append(fpm, ver)
			}
		}
	}
	if len(fpm) == 0 {
		fmt.Println("php-fpm: down")
	} else {
		fmt.Printf("php-fpm: up (%s)\n", strings.Join(fpm, ", "))
	}
	// Shares (cloudflared tunnels)
	entries, _ = os.ReadDir(filepath.Join(xpierHome(), "servers"))
	shares := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "share-") && strings.HasSuffix(e.Name(), ".json") {
			site := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "share-"), ".json")
			if st, err := loadShareState(site); err == nil && pidAlive(st.PID) {
				shares = append(shares, fmt.Sprintf("%s(%s)", site, st.URL))
			}
		}
	}
	if len(shares) == 0 {
		fmt.Println("share:   none (cloudflared)")
	} else {
		fmt.Printf("share:   up: %s\n", strings.Join(shares, ", "))
	}
}

func cmdServices(args []string) error {
	xpierServiceStatus()
	return nil
}

func stopDaemon(label string) error {
	out, err := runOut("launchctl", "bootout", "system/"+label)
	if err != nil {
		return fmt.Errorf("bootout %s: %v: %s", label, err, out)
	}
	return nil
}

func startDaemon(label string) error {
	return launchctlBootstrap(label, filepath.Join(launchdDir(), label+".plist"))
}

func restartDaemon(label string) error {
	runOutErr("launchctl", "bootout", "system/"+label)
	return startDaemon(label)
}

func stopDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := stopDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

func startDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := startDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

// showConfig prints the path and opens the file in an editor when available,
// otherwise prints the content.
func showConfig(path string) error {
	fmt.Printf("config: %s\n", path)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if fileExists("/usr/local/bin/code") {
			editor = "/usr/local/bin/code"
		}
	}
	if editor != "" {
		return runOutErr(editor, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

// cmdService manages a single service: nginx / dnsmasq / php-fpm[-<ver>].
func cmdService(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: xpier service <nginx|dnsmasq|php-fpm|php-fpm-8.2> <status|config|configtest|reload|start|stop|restart>")
	}
	name, action := args[0], args[1]
	ver := ""
	if strings.HasPrefix(name, "php-fpm") {
		ver = strings.TrimPrefix(name, "php-fpm-")
		if ver == name || ver == "" {
			ver = defaultPhpVersion()
		}
		name = "php-fpm"
	}
	switch name {
	case "nginx":
		switch action {
		case "status":
			b, _ := portBusy("80")
			fmt.Printf("nginx: %s\n", upDown(b))
		case "config":
			return showConfig(filepath.Join(nginxHome(), "nginx.conf"))
		case "configtest":
			out, err := runOut("sudo", "-n", nginxBin(), "-t", "-c", filepath.Join(nginxHome(), "nginx.conf"))
			if err != nil {
				return fmt.Errorf("nginx configtest failed: %v: %s", err, out)
			}
			fmt.Println("nginx configuration test successful")
		case "reload":
			if err := nginxReload(); err != nil {
				return err
			}
			fmt.Println("nginx reloaded")
		case "start", "stop", "restart":
			if os.Geteuid() != 0 {
				return fmt.Errorf("nginx is a launchd daemon; run `sudo xpier service nginx %s`", action)
			}
			var err error
			switch action {
			case "start":
				err = startDaemon("com.xpier.nginx")
			case "stop":
				err = stopDaemon("com.xpier.nginx")
			case "restart":
				err = restartDaemon("com.xpier.nginx")
			}
			if err != nil {
				return err
			}
			fmt.Printf("nginx %s\n", action)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	case "dnsmasq":
		switch action {
		case "status":
			fmt.Printf("dnsmasq: %s\n", upDown(dnsmasqRunning()))
		case "config":
			return showConfig(dnsmasqConfPath())
		case "configtest":
			out, err := runOut(dnsmasqBin(), "--test", "-C", dnsmasqConfPath())
			if err != nil {
				return fmt.Errorf("dnsmasq configtest failed: %v: %s", err, out)
			}
			fmt.Println("dnsmasq configuration test successful")
		case "start", "stop", "restart":
			if os.Geteuid() != 0 {
				return fmt.Errorf("dnsmasq is a launchd daemon; run `sudo xpier service dnsmasq %s`", action)
			}
			var err error
			switch action {
			case "start":
				err = startDaemon("com.xpier.dnsmasq")
			case "stop":
				err = stopDaemon("com.xpier.dnsmasq")
			case "restart":
				err = restartDaemon("com.xpier.dnsmasq")
			}
			if err != nil {
				return err
			}
			fmt.Printf("dnsmasq %s\n", action)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	case "php-fpm":
		switch action {
		case "status":
			fmt.Printf("php-fpm %s: %s\n", ver, upDown(fpmRunning(ver)))
		case "config":
			return showConfig(fpmConfPath(ver))
		case "start":
			return fpmUp(ver)
		case "stop":
			return fpmDown(ver)
		case "restart":
			fpmDown(ver)
			return fpmUp(ver)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	default:
		return fmt.Errorf("unknown service %q (nginx|dnsmasq|php-fpm)", name)
	}
	return nil
}

func cmdServicesStop(args []string) error {
	entries, _ := os.ReadDir(filepath.Join(xpierHome(), "servers"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			fpmDown(ver)
		}
	}
	if os.Geteuid() == 0 {
		return stopDaemons()
	}
	fmt.Println("php-fpm stopped; nginx/dnsmasq still running (run `sudo xpier services:stop` to stop them too)")
	return nil
}

func cmdServicesStart(args []string) error {
	if os.Geteuid() == 0 {
		if err := startDaemons(); err != nil {
			return err
		}
		fmt.Println("nginx + dnsmasq daemons started")
	} else {
		fmt.Println("starting php-fpm for linked sites (run `sudo xpier services:start` to also start nginx/dnsmasq)")
	}
	return cmdSitesUp(args)
}
