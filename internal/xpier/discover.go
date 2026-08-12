package xpier

import (
	"regexp"
	"strconv"
	"strings"
	"xpier/internal/store"
)

var (
	versionRe = regexp.MustCompile(`PHP\s+(\d+\.\d+\.\d+)`)
	extVerRe  = regexp.MustCompile(`Version\s*=>\s*(\S+)`)
)

func phpBinFor(version string) string {
	return store.BrewPrefix() + "/opt/php@" + version + "/bin/php"
}

func phpCandidates(version string) []string {
	return []string{
		phpBinFor(version),
		"/usr/local/bin/php" + version,
	}
}

func phpVersion(bin string) string {
	out, err := store.RunOut(bin, "-v")
	if err != nil {
		return ""
	}
	if m := versionRe.FindStringSubmatch(out); len(m) > 1 {
		return m[1]
	}
	return ""
}

func extVersion(bin, ext string) string {
	out, err := store.RunOut(bin, "--ri", ext)
	if err != nil {
		return ""
	}
	if m := extVerRe.FindStringSubmatch(out); len(m) > 1 {
		return m[1]
	}
	return ""
}

func extLoaded(bin, ext string) bool {
	_, err := store.RunOut(bin, "--ri", ext)
	return err == nil
}

type verTuple struct{ major, minor, patch int }

func parseVer(s string) (verTuple, bool) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) < 2 {
		return verTuple{}, false
	}
	t := verTuple{}
	t.major, _ = strconv.Atoi(parts[0])
	t.minor, _ = strconv.Atoi(parts[1])
	if len(parts) > 2 {
		t.patch, _ = strconv.Atoi(parts[2])
	}
	return t, true
}

func compareVer(a, b verTuple) int {
	switch {
	case a.major != b.major:
		return a.major - b.major
	case a.minor != b.minor:
		return a.minor - b.minor
	default:
		return a.patch - b.patch
	}
}

// constraintOk supports "*", exact, ">=x.y", "^x.y".
func constraintOk(constraint, installed string) bool {
	c := strings.TrimSpace(constraint)
	if c == "" || c == "*" {
		return true
	}
	iv, ok := parseVer(installed)
	if !ok {
		return false
	}
	var cv verTuple
	switch {
	case strings.HasPrefix(c, ">="):
		cv, _ = parseVer(strings.TrimPrefix(c, ">="))
		return compareVer(iv, cv) >= 0
	case strings.HasPrefix(c, "^"):
		cv, _ = parseVer(strings.TrimPrefix(c, "^"))
		return iv.major == cv.major && compareVer(iv, cv) >= 0
	default:
		cv, _ = parseVer(c)
		return compareVer(iv, cv) == 0
	}
}

func serviceRunning(name string) bool {
	out, err := store.RunOut("brew", "services", "list")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == name && fields[1] == "started" {
			return true
		}
	}
	return false
}
