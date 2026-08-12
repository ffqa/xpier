package xpier

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xpier/internal/store"
)

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
	if out, err := brewAsUser("list", "--versions", formula); err == nil && strings.Contains(out, formula) {
		fmt.Printf("%s already installed\n", formula)
	} else {
		fmt.Printf("installing %s via brew...\n", formula)
		if out, err := brewAsUser("install", formula); err != nil {
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
	if out, err := brewAsUser("list", "--versions", formula); err != nil || !strings.Contains(out, formula) {
		ok, err := store.ConfirmYesNo(fmt.Sprintf("%s 未安装（brew install %s），是否现在安装？", formula, formula))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%s not installed; run `xpier db:install %s` first", formula, fs.Arg(0))
		}
		fmt.Printf("installing %s...\n", formula)
		if out, err := brewAsUser("install", formula); err != nil {
			return fmt.Errorf("brew install %s: %v: %s", formula, err, out)
		}
	}
	if out, err := brewAsUser("services", "start", formula); err != nil {
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
	if out, err := brewAsUser("services", "stop", formula); err != nil {
		return fmt.Errorf("brew services stop %s: %v: %s", formula, err, out)
	}
	fmt.Printf("%s stopped\n", formula)
	return nil
}

func cmdDBCreate(args []string) error {
	fs := flag.NewFlagSet("db:create", flag.ExitOnError)
	svc := fs.String("db", "mysql", "database type (mysql|mariadb|postgres)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: xpier db:create <name> [--db mysql]")
	}
	name := fs.Arg(0)
	if !safeSiteNameRe.MatchString(name) {
		return fmt.Errorf("invalid database name %q", name)
	}
	switch *svc {
	case "mysql", "mariadb":
		// MySQL creates the DB with the user's privileges; default brew mysql
		// root has no password locally.
		out, err := exec.Command("mysql", "-u", "root", "-e", "CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4;").CombinedOutput()
		if err != nil {
			return fmt.Errorf("mysql create database: %v: %s", err, out)
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

// adminerSite registers a `database.test` site serving Adminer.
func ensureAdminerSite() error {
	dir := filepath.Join(store.XpierHome(), "adminer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	index := filepath.Join(dir, "index.php")
	if !store.FileExists(index) {
		// Download the single-file Adminer (MIT).
		if out, err := store.RunOut("curl", "-fsSL", "-o", index, "https://www.adminer.org/latest-mysql-en.php"); err != nil || !store.FileExists(index) {
			return fmt.Errorf("download adminer: %v: %s", err, out)
		}
	}
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	if _, ok := sites.Sites["database"]; !ok {
		sites.Sites["database"] = store.Site{Path: dir, Driver: "php"}
		if err := sites.Save(); err != nil {
			return err
		}
	}
	return writeSiteNginxConfig(sites, "database")
}

func cmdDB(args []string) error {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureAdminerSite(); err != nil {
		return err
	}
	siteName := "database"
	if fs.NArg() > 0 {
		siteName = fs.Arg(0) // open a specific site's db via adminer
	}
	_ = siteName
	return runOutErr("open", "http://database.test/")
}
