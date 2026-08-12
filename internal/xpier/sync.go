package xpier

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"xpier/internal/store"
)

var (
	safeExtRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	// safeSvcRe keeps manifest service names out of shell command injection
	// (brew install <svc> is passed through sh -c).
	safeSvcRe = regexp.MustCompile(`^[a-zA-Z0-9@._-]+$`)
)

type planItem struct {
	kind    string // php | ext | svc
	name    string
	state   string // ok | missing
	detail  string
	command string
}

func cmdSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	apply := fs.Bool("apply", false, "actually run install commands and write xpier.lock")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath, lockPath := store.ResolvePaths(cwd)
	m, err := store.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("load %s: %w (create one with `xpier init --php 8.2`)", manifestPath, err)
	}
	items := plan(m)
	pending := 0
	for _, it := range items {
		fmt.Printf("[%s] %-4s %-10s %s\n", it.state, it.kind, it.name, it.detail)
		if it.command != "" {
			fmt.Printf("       would run: %s\n", it.command)
		}
		if it.state != "ok" {
			pending++
		}
	}
	if pending == 0 {
		fmt.Println(store.Paint("nothing to do"))
		return writeLock(m, lockPath)
	}
	if !*apply {
		fmt.Printf("dry-run: %d item(s) pending; re-run with --apply to install and write %s\n", pending, lockPath)
		return nil
	}
	if err := applyPlan(items); err != nil {
		return err
	}
	return writeLock(m, lockPath)
}

func plan(m *store.Manifest) []planItem {
	var items []planItem
	phpOK := false
	bin := phpBinFor(m.PHP)
	if v := phpVersion(bin); v != "" {
		items = append(items, planItem{"php", m.PHP, "ok", bin + " (" + v + ")", ""})
		phpOK = true
	} else if out, err := store.RunOut("which", "php"); err == nil && out != "" && strings.HasPrefix(phpVersion(out), m.PHP+".") {
		items = append(items, planItem{"php", m.PHP, "ok", "using " + out + " (" + phpVersion(out) + ")", ""})
		phpOK = true
	} else {
		cmd := "brew tap shivammathur/php && brew install shivammathur/php/php@" + m.PHP
		if !store.SafePhpRe.MatchString(m.PHP) {
			cmd = ""
		}
		items = append(items, planItem{"php", m.PHP, "missing", "php@" + m.PHP + " not found", cmd})
	}
	for _, ext := range store.SortedKeys(m.Extensions) {
		it := planItem{kind: "ext", name: ext}
		if !phpOK {
			it.state = "missing"
			it.detail = "requires php@" + m.PHP + " first"
			items = append(items, it)
			continue
		}
		if v := extVersion(bin, ext); v != "" {
			if constraintOk(m.Extensions[ext], v) {
				it.state = "ok"
				it.detail = v + " satisfies " + m.Extensions[ext]
			} else {
				it.state = "missing"
				it.detail = "installed " + v + " does not satisfy " + m.Extensions[ext]
			}
		} else {
			it.state = "missing"
			it.detail = "not loaded in php@" + m.PHP
			if safeExtRe.MatchString(ext) && store.SafePhpRe.MatchString(m.PHP) {
				it.command = "brew tap shivammathur/extensions && brew install shivammathur/extensions/" + ext + "@" + m.PHP
			}
		}
		items = append(items, it)
	}
	for _, svc := range m.Services {
		it := planItem{kind: "svc", name: svc}
		if out, err := store.RunOut("brew", "list", "--versions", svc); err != nil {
			it.state = "missing"
			it.detail = "not installed via brew"
			it.command = "brew install " + svc
			if !safeSvcRe.MatchString(svc) {
				it.command = ""
				it.detail = "invalid service name"
			}
		} else {
			it.state = "ok"
			it.detail = out
		}
		items = append(items, it)
	}
	return items
}

func applyPlan(items []planItem) error {
	for _, it := range items {
		if it.state == "ok" || it.command == "" {
			continue
		}
		fmt.Println("-> " + it.command)
		switch {
		case strings.Contains(it.command, "shivammathur/extensions"):
			store.BrewTrustTap("shivammathur/extensions")
		case strings.Contains(it.command, "shivammathur/php"):
			store.BrewTrustTap("shivammathur/php")
		}
		if err := store.RunOutLive("sh", "-c", it.command); err != nil {
			return fmt.Errorf("%s failed: %w", it.command, err)
		}
	}
	return nil
}

func writeLock(m *store.Manifest, lockPath string) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	bin := phpBinFor(m.PHP)
	ver := phpVersion(bin)
	path := bin
	if ver == "" {
		if out, err := store.RunOut("which", "php"); err == nil && out != "" {
			if v := phpVersion(out); v != "" {
				bin, ver, path = out, v, out
			}
		}
	}
	lock := store.Lock{
		SchemaVersion: 1,
		GeneratedAt:   time.Now().Format(time.RFC3339),
		PHP:           store.PhpLock{Version: ver, Path: path},
	}
	for _, ext := range store.SortedKeys(m.Extensions) {
		lock.Extensions = append(lock.Extensions, store.ExtLock{
			Name:       ext,
			Constraint: m.Extensions[ext],
			Installed:  extVersion(bin, ext),
			Loaded:     extLoaded(bin, ext),
		})
	}
	for _, svc := range m.Services {
		lock.Services = append(lock.Services, store.ServiceLock{Name: svc, Running: serviceRunning(svc)})
	}
	if err := lock.Save(lockPath); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", lockPath)
	return nil
}
