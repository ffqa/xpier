package xpier

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"xpier/internal/store"
)

func cmdDoctor(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath, lockPath := store.ResolvePaths(cwd)
	m, mErr := store.LoadManifest(manifestPath)
	lock, lockErr := store.LoadLock(lockPath)
	failed := false

	// The manifest is optional: without it, doctor checks what can be
	// auto-detected and still runs the composer platform check.
	hasManifest := mErr == nil
	expectPHP := ""
	extConstraints := map[string]string{}
	if hasManifest {
		expectPHP = m.PHP
		extConstraints = m.Extensions
		fmt.Printf("pins: php %s, runtime %s", m.PHP, m.Runtime)
		for _, ext := range store.SortedKeys(m.Extensions) {
			fmt.Printf(", %s %s", ext, m.Extensions[ext])
		}
		fmt.Println()
		if lockErr == nil {
			fmt.Printf("lock: php %s @ %s\n", lock.PHP.Version, lock.PHP.Path)
		}
	} else {
		fmt.Println("[info] no manifest; checking auto-detected environment (create one with `xpier init --php 8.2` to pin versions)")
	}

	bin := ""
	if lockErr == nil && lock.PHP.Path != "" && store.FileExists(lock.PHP.Path) {
		bin = lock.PHP.Path
	} else if hasManifest && m.PHP != "" {
		for _, c := range phpCandidates(m.PHP) {
			if store.FileExists(c) {
				bin = c
				break
			}
		}
	}
	if bin == "" {
		if out, err := store.RunOut("which", "php"); err == nil && out != "" {
			bin = out
		}
	}

	// php
	if bin == "" {
		failed = true
		fmt.Println("[fail] php     no php binary found")
	} else {
		v := phpVersion(bin)
		if expectPHP != "" && (v == "" || !versionPrefixMatch(expectPHP, v)) {
			failed = true
			fmt.Printf("[fail] php     expected %s, found %s (%s)\n", expectPHP, v, bin)
		} else {
			note := ""
			if lockErr == nil && lock.PHP.Version != "" && lock.PHP.Version != v {
				note = fmt.Sprintf(" (drift: lock recorded %s)", lock.PHP.Version)
			}
			fmt.Printf("[ok]   php     %s @ %s%s\n", v, bin, note)
		}
	}

	// extensions
	for _, ext := range store.SortedKeys(extConstraints) {
		if bin == "" {
			failed = true
			fmt.Printf("[fail] ext     %-10s php binary unavailable\n", ext)
			continue
		}
		v := extVersion(bin, ext)
		if v == "" {
			failed = true
			fmt.Printf("[fail] ext     %-10s not loaded in %s\n", ext, bin)
			continue
		}
		if !constraintOk(extConstraints[ext], v) {
			failed = true
			fmt.Printf("[fail] ext     %-10s %s does not satisfy %s\n", ext, v, extConstraints[ext])
			continue
		}
		note := ""
		if lockErr == nil {
			for _, e := range lock.Extensions {
				if e.Name == ext && e.Installed != "" && e.Installed != v {
					note = fmt.Sprintf(" (drift: lock recorded %s)", e.Installed)
				}
			}
		}
		fmt.Printf("[ok]   ext     %-10s %s satisfies %s%s\n", ext, v, extConstraints[ext], note)
	}

	// services
	if hasManifest {
		for _, svc := range m.Services {
			if serviceRunning(svc) {
				fmt.Printf("[ok]   svc     %-10s running\n", svc)
			} else {
				fmt.Printf("[warn] svc     %-10s not running (brew services start %s)\n", svc, svc)
			}
		}
	}

	// composer platform check
	if store.FileExists("composer.lock") {
		if err := checkComposerPlatformReqs(); err != nil {
			failed = true
		}
	} else {
		fmt.Println("[info] composer.lock absent, skipping platform check")
	}

	if failed {
		return fmt.Errorf("doctor found problems")
	}
	return nil
}

func versionPrefixMatch(expect, actual string) bool {
	if expect == actual {
		return true
	}
	return strings.HasPrefix(actual, expect+".")
}

func checkComposerPlatformReqs() error {
	composer, err := exec.LookPath("composer")
	if err != nil {
		fmt.Println("[warn] composer not on PATH, skipping platform check")
		return nil
	}
	out, err := exec.Command(composer, "--ansi", "check-platform-reqs", "--no-dev").CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		fmt.Printf("[fail] composer check-platform-reqs failed\n")
		return err
	}
	fmt.Println("[ok]   composer platform requirements satisfied")
	return nil
}
