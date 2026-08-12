package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/share"
	"xpier/internal/store"
)

func DnsmasqRunning() bool {
	// lsof sometimes fails to report dnsmasq's UDP 53 socket on macOS;
	// detect the process instead.
	out, err := store.RunOut("pgrep", "-f", DnsmasqBin())
	return err == nil && strings.TrimSpace(out) != ""
}

func XpierServiceStatus() {
	nginxUp := false
	if b, _ := store.PortBusy("80"); b {
		nginxUp = true
	}
	dnsUp := DnsmasqRunning()
	fmt.Printf("nginx:   %s\n", store.UpDown(nginxUp))
	fmt.Printf("dnsmasq: %s\n", store.UpDown(dnsUp))
	entries, _ := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	fpm := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			if FpmRunning(ver) {
				fpm = append(fpm, ver)
			}
		}
	}
	if len(fpm) == 0 {
		fmt.Println("php-fpm: " + store.Paint("down"))
	} else {
		fmt.Printf("php-fpm: %s (%s)\n", store.Paint("up"), strings.Join(fpm, ", "))
	}
	// Shares (cloudflared tunnels)
	entries, _ = os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	shares := []string{}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "share-") && strings.HasSuffix(e.Name(), ".json") {
			site := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "share-"), ".json")
			if st, err := share.LoadShareState(site); err == nil && store.PidAlive(st.PID) {
				shares = append(shares, fmt.Sprintf("%s(%s)", site, st.URL))
			}
		}
	}
	if len(shares) == 0 {
		fmt.Println("share:   " + store.Paint("none") + " (cloudflared)")
	} else {
		fmt.Printf("share:   %s: %s\n", store.Paint("up"), strings.Join(shares, ", "))
	}
}

func CmdServices(args []string) error {
	XpierServiceStatus()
	return nil
}

func StopDaemon(label string) error {
	out, err := store.RunOut("launchctl", "bootout", "system/"+label)
	if err != nil {
		return fmt.Errorf("bootout %s: %v: %s", label, err, out)
	}
	return nil
}

func StartDaemon(label string) error {
	return LaunchctlBootstrap(label, filepath.Join(LaunchdDir(), label+".plist"))
}

func RestartDaemon(label string) error {
	store.RunOutErr("launchctl", "bootout", "system/"+label)
	return StartDaemon(label)
}

func StopDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := StopDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

func StartDaemons() error {
	for _, label := range []string{"com.xpier.nginx", "com.xpier.dnsmasq"} {
		if err := StartDaemon(label); err != nil {
			return err
		}
	}
	return nil
}

// ShowConfig prints the path and opens the file in an editor when available,
// otherwise prints the content.
func ShowConfig(path string) error {
	fmt.Printf("config: %s\n", path)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if store.FileExists("/usr/local/bin/code") {
			editor = "/usr/local/bin/code"
		}
	}
	if editor != "" {
		return store.RunOutErr(editor, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

// CmdService manages a single service: nginx / dnsmasq / php-fpm[-<ver>].
func CmdService(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: xpier service <nginx|dnsmasq|php-fpm|php-fpm-8.2> <status|config|configtest|reload|start|stop|restart>")
	}
	name, action := args[0], args[1]
	ver := ""
	if strings.HasPrefix(name, "php-fpm") {
		ver = strings.TrimPrefix(name, "php-fpm-")
		if ver == name || ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
		name = "php-fpm"
	}
	switch name {
	case "nginx":
		switch action {
		case "status":
			b, _ := store.PortBusy("80")
			fmt.Printf("nginx: %s\n", store.UpDown(b))
		case "config":
			return ShowConfig(filepath.Join(nginx.NginxHome(), "nginx.conf"))
		case "configtest":
			out, err := store.RunOut("sudo", "-n", nginx.NginxBin(), "-t", "-c", filepath.Join(nginx.NginxHome(), "nginx.conf"))
			if err != nil {
				return fmt.Errorf("nginx configtest failed: %v: %s", err, out)
			}
			fmt.Println("nginx configuration test successful")
		case "reload":
			if err := nginx.NginxReload(); err != nil {
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
				err = StartDaemon("com.xpier.nginx")
			case "stop":
				err = StopDaemon("com.xpier.nginx")
			case "restart":
				err = RestartDaemon("com.xpier.nginx")
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
			fmt.Printf("dnsmasq: %s\n", store.UpDown(DnsmasqRunning()))
		case "config":
			return ShowConfig(store.DnsmasqConfPath())
		case "configtest":
			out, err := store.RunOut(DnsmasqBin(), "--test", "-C", store.DnsmasqConfPath())
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
				err = StartDaemon("com.xpier.dnsmasq")
			case "stop":
				err = StopDaemon("com.xpier.dnsmasq")
			case "restart":
				err = RestartDaemon("com.xpier.dnsmasq")
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
			fmt.Printf("php-fpm %s: %s\n", ver, store.UpDown(FpmRunning(ver)))
		case "config":
			return ShowConfig(FpmConfPath(ver))
		case "start":
			return FpmUp(ver)
		case "stop":
			return FpmDown(ver)
		case "restart":
			FpmDown(ver)
			return FpmUp(ver)
		default:
			return fmt.Errorf("unknown action %q", action)
		}
	default:
		return fmt.Errorf("unknown service %q (nginx|dnsmasq|php-fpm)", name)
	}
	return nil
}

func CmdServicesStop(args []string) error {
	entries, _ := os.ReadDir(filepath.Join(store.XpierHome(), "servers"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "fpm-") && strings.HasSuffix(e.Name(), ".json") {
			ver := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "fpm-"), ".json")
			FpmDown(ver)
		}
	}
	if os.Geteuid() == 0 {
		return StopDaemons()
	}
	fmt.Println("php-fpm stopped; nginx/dnsmasq still running (run `sudo xpier services:stop` to stop them too)")
	return nil
}

func CmdServicesStart(args []string) error {
	if os.Geteuid() == 0 {
		if err := StartDaemons(); err != nil {
			return err
		}
		fmt.Println("nginx + dnsmasq daemons started")
	} else {
		fmt.Println("starting php-fpm for linked sites (run `sudo xpier services:start` to also start nginx/dnsmasq)")
	}
	return startSiteFpms()
}

func startSiteFpms() error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	versions := map[string]bool{}
	for _, site := range sites.Sites {
		ver := site.PHP
		if ver == "" {
			ver = nginx.DefaultPhpVersion()
		}
		versions[ver] = true
	}
	for ver := range versions {
		if err := FpmUp(ver); err != nil {
			fmt.Printf("[warn] %v\n", err)
		}
	}
	return nil
}
