package xpier

import (
	_ "embed"

	"bytes"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/service"
	"xpier/internal/store"
)

//go:embed adminer.php
var adminerPHP []byte

var dbServices = map[string]string{
	"mysql":      "mysql",
	"maria":      "mariadb",
	"redis":      "redis",
	"postgres":   "postgresql@16",
	"pgsql":      "postgresql@16",
	"postgresql": "postgresql@16",
}

// dbFormula maps a service name to its brew formula.
func dbFormula(svc string) (string, error) {
	if f, ok := dbServices[svc]; ok {
		return f, nil
	}
	return "", fmt.Errorf("unknown database %q (mysql|mariadb|redis|postgres)", svc)
}

func cmdDBInstall(args []string) error {
	fs := flag.NewFlagSet("db:install", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier db:install <mysql|mariadb|redis|postgres>")
	}
	svc := fs.Arg(0)
	formula, err := dbFormula(svc)
	if err != nil {
		return err
	}
	if out, err := service.BrewAsUser("list", "--versions", formula); err == nil && strings.Contains(out, formula) {
		fmt.Printf("%s already installed\n", formula)
	} else {
		fmt.Printf("installing %s via brew...\n", formula)
		if out, err := service.BrewAsUser("install", formula); err != nil {
			return fmt.Errorf("brew install %s: %v: %s", formula, err, out)
		}
	}
	return nil
}

func cmdDBStart(args []string) error {
	fs := flag.NewFlagSet("db:start", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier db:start <mysql|mariadb|redis|postgres>")
	}
	formula, err := dbFormula(fs.Arg(0))
	if err != nil {
		return err
	}
	if out, err := service.BrewAsUser("list", "--versions", formula); err != nil || !strings.Contains(out, formula) {
		ok, err := store.ConfirmYesNo(fmt.Sprintf("%s 未安装（brew install %s），是否现在安装？", formula, formula))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s not installed; run `xpier db:install %s` first", formula, fs.Arg(0))
		}
		fmt.Printf("installing %s...\n", formula)
		if out, err := service.BrewAsUser("install", formula); err != nil {
			return fmt.Errorf("brew install %s: %v: %s", formula, err, out)
		}
	}
	if out, err := service.BrewAsUser("services", "start", formula); err != nil {
		return fmt.Errorf("brew services start %s: %v: %s", formula, err, out)
	}
	fmt.Printf("%s started (brew services)\n", formula)
	return nil
}

func cmdDBStop(args []string) error {
	fs := flag.NewFlagSet("db:stop", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier db:stop <mysql|mariadb|redis|postgres>")
	}
	formula, err := dbFormula(fs.Arg(0))
	if err != nil {
		return err
	}
	if out, err := service.BrewAsUser("services", "stop", formula); err != nil {
		return fmt.Errorf("brew services stop %s: %v: %s", formula, err, out)
	}
	fmt.Printf("%s stopped\n", formula)
	return nil
}

func cmdDBCreate(args []string) error {
	fs := flag.NewFlagSet("db:create", flag.ExitOnError)
	svc := fs.String("db", "mysql", "database type (mysql|mariadb|postgres)")
	user := fs.String("user", "root", "database user")
	password := fs.String("password", "", "database password (empty = no password / socket auth)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier db:create <name> [--db mysql] [--user root] [--password ...]")
	}
	name := fs.Arg(0)
	if !store.SafeSiteNameRe.MatchString(name) {
		return fmt.Errorf("invalid database name %q", name)
	}
	switch *svc {
	case "mysql", "mariadb":
		// Default brew MySQL/MariaDB authenticate via unix socket with an
		// empty root password; --user/--password cover password-protected
		// setups instead of silently failing on `mysql -u root`.
		mysqlArgs := []string{}
		if *user != "" {
			mysqlArgs = append(mysqlArgs, "-u", *user)
		}
		if *password != "" {
			mysqlArgs = append(mysqlArgs, "-p"+*password)
		}
		mysqlArgs = append(mysqlArgs, "-e", "CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4;")
		out, err := exec.Command("mysql", mysqlArgs...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("mysql create database: %v: %s (root 有密码?用 `xpier db:create %s --user root --password ...` 或先 `mysql -u root -p` 设置好凭据)", err, out, name)
		}
		fmt.Printf("created database %s (mysql)\n", name)
	case "postgres", "pgsql", "postgresql":
		bin := filepath.Join(store.BrewPrefix(), "opt", "postgresql@16", "bin", "createdb")
		out, err := exec.Command(bin, name).CombinedOutput()
		if err != nil {
			return fmt.Errorf("createdb: %v: %s", err, out)
		}
		fmt.Printf("created database %s (postgres)\n", name)
	default:
		return fmt.Errorf("db:create supports mysql/mariadb/postgres only")
	}
	return nil
}

// adminerSite registers a `database.<tld>` site serving Adminer.
func ensureAdminerSite() error {
	dir := filepath.Join(store.XpierHome(), "adminer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	index := filepath.Join(dir, "index.php")
	// Adminer is embedded in the binary (Apache-2.0 / GPL-2, see the file
	// header) so `xpier db` works offline. Rewrite whenever the deployed
	// copy differs (e.g. after an embedded patch like the empty-password one).
	if data, err := os.ReadFile(index); err != nil || !bytes.Equal(data, adminerPHP) {
		if err := os.WriteFile(index, adminerPHP, 0o644); err != nil {
			return fmt.Errorf("write adminer: %w", err)
		}
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if existing, ok := sites.Sites["database"]; ok && existing.Path != dir {
		return fmt.Errorf("a user site named `database` already exists at %s; xpier reserves the name for Adminer", existing.Path)
	}
	sites.Sites["database"] = store.Site{Path: dir, Driver: "php"}
	if err := sites.Save(); err != nil {
		return err
	}
	if err := nginx.WriteSiteNginxConfig(sites, "database"); err != nil {
		return err
	}
	// The config only takes effect after a reload; without this the browser
	// opens database.<tld> to a 404 (default server).
	if err := nginx.NginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v (run `sudo xpier install` to fix sudoers, then `xpier db` again)\n", err)
	}
	return nil
}

// detectMySQLServer inspects the running mysqld (brew, DBngin, ...) and
// returns a "host:port" Adminer can use. TCP is preferred because PHP's
// mysqli socket path often differs from the server's (e.g. DBngin uses
// /tmp/mysql_3306.sock while PHP defaults to /tmp/mysql.sock).
func detectMySQLServer() string {
	out, err := store.RunOut("pgrep", "-f", "mysqld")
	if err != nil {
		return ""
	}
	pid := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
	if pid == "" {
		return ""
	}
	cmdline, err := store.RunOut("ps", "-o", "command=", "-p", pid)
	if err != nil {
		return ""
	}
	port := "3306"
	for _, arg := range strings.Fields(cmdline) {
		if strings.HasPrefix(arg, "--port=") {
			port = strings.TrimPrefix(arg, "--port=")
		}
	}
	return "127.0.0.1:" + port
}

// adminerURL builds the Adminer URL. siteName prefills the database field
// (Laravel convention: database name == site name); server prefills the
// server field so the user does not have to fight socket paths.
func adminerURL(sites *store.Sites, siteName, server string) string {
	u := "http://" + store.SiteDomain(sites, "database") + "/"
	var params []string
	if siteName != "" {
		params = append(params, "db="+siteName)
	}
	if server != "" {
		params = append(params, "server="+url.QueryEscape(server), "username=root")
	}
	if len(params) > 0 {
		u += "?" + strings.Join(params, "&")
	}
	return u
}

func cmdDB(args []string) error {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureAdminerSite(); err != nil {
		return err
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	siteName := ""
	if fs.NArg() > 0 {
		siteName = fs.Arg(0)
	}
	return store.RunOutErr("open", adminerURL(sites, siteName, detectMySQLServer()))
}
