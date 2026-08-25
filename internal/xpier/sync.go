package xpier

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"xpier/internal/service"
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
	// argv is the arg-vector form of the install command (e.g. ["brew","install",svc]).
	// Stored instead of a shell string so applyPlan never touches sh -c; the
	// human-readable summary is rebuilt from it for the dry-run preview.
	argv []string
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
		return fmt.Errorf("load %s: %w (create one with `xpier env:init --php 8.2`)", manifestPath, err)
	}
	items := plan(m)
	pending := 0
	for _, it := range items {
		fmt.Printf("[%s] %-4s %-10s %s\n", it.state, it.kind, it.name, it.detail)
		if len(it.argv) > 0 {
			fmt.Printf("       would run: %s\n", strings.Join(it.argv, " "))
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
	// --apply means "go install": trust the standard PHP taps up front
	// (Homebrew refuses untrusted tap formulae otherwise).
	for _, tap := range []string{"shivammathur/php", "shivammathur/extensions"} {
		if err := store.BrewTrustTap(tap); err != nil {
			fmt.Printf("[warn] brew trust %s: %v\n", tap, err)
		}
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
		items = append(items, planItem{"php", m.PHP, "ok", bin + " (" + v + ")", nil})
		phpOK = true
	} else if out, err := store.RunOut("which", "php"); err == nil && out != "" && strings.HasPrefix(phpVersion(out), m.PHP+".") {
		items = append(items, planItem{"php", m.PHP, "ok", "using " + out + " (" + phpVersion(out) + ")", nil})
		phpOK = true
	} else {
		var argv []string
		if store.SafePhpRe.MatchString(m.PHP) {
			argv = []string{"brew", "tap", "shivammathur/php", "&&", "brew", "install", "shivammathur/php/php@" + m.PHP}
		}
		items = append(items, planItem{"php", m.PHP, "missing", "php@" + m.PHP + " not found", argv})
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
				it.argv = []string{"brew", "tap", "shivammathur/extensions", "&&", "brew", "install", "shivammathur/extensions/" + ext + "@" + m.PHP}
			}
		}
		items = append(items, it)
	}
	for _, svc := range m.Services {
		it := planItem{kind: "svc", name: svc}
		if out, err := service.BrewAsUser("list", "--versions", svc); err != nil {
			it.state = "missing"
			it.detail = "not installed via brew"
			if safeSvcRe.MatchString(svc) {
				it.argv = []string{"brew", "install", svc}
			} else {
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
		if it.state == "ok" || len(it.argv) == 0 {
			continue
		}
		fmt.Println("-> " + strings.Join(it.argv, " "))
		// Re-trust the tap that the corresponding plan branch validated; safe
		// Homebrew installs require a trusted tap up front.
		switch it.kind {
		case "ext":
			store.BrewTrustTap("shivammathur/extensions")
		case "php":
			store.BrewTrustTap("shivammathur/php")
		}
		// argv may contain "&&" to chain steps (tap then install); run each
		// sub-command directly via arg-vector, never through sh -c.
		for _, sub := range splitChain(it.argv) {
			if err := runLive(sub); err != nil {
				return fmt.Errorf("%s failed: %w", strings.Join(sub, " "), err)
			}
		}
	}
	return nil
}

// runLive runs a planned sub-command with live output. brew is routed through
// service.BrewAsUserLive so installs work under sudo (Homebrew refuses root);
// any other command falls back to a direct live exec.
func runLive(argv []string) error {
	if len(argv) > 0 && argv[0] == "brew" {
		return service.BrewAsUserLive(argv[1:]...)
	}
	return store.RunOutLiveYes(argv[0], argv[1:]...)
}

// splitChain splits an argv on a literal "&&" token into sub-commands, so a
// plan item like ["brew","tap",X,"&&","brew","install",Y] becomes two
// arg-vectors. Lets plan() express multi-step installs as one item without
// applyPlan ever touching a shell.
func splitChain(argv []string) [][]string {
	var groups [][]string
	cur := make([]string, 0, len(argv))
	for _, a := range argv {
		if a == "&&" {
			if len(cur) > 0 {
				groups = append(groups, cur)
				cur = make([]string, 0, len(argv))
			}
			continue
		}
		cur = append(cur, a)
	}
	if len(cur) > 0 {
		groups = append(groups, cur)
	}
	return groups
}

func writeLock(m *store.Manifest, lockPath string) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	// Drop the pre-0.0.68 xpier.lock name (hidden .xpier.lock is canonical).
	os.Remove(filepath.Join(filepath.Dir(lockPath), store.LegacyLockName))
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
