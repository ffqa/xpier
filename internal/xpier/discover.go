package xpier

import (
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	versionRe = regexp.MustCompile(`PHP\s+(\d+\.\d+\.\d+)`)
	extVerRe  = regexp.MustCompile(`Version\s*=>\s*(\S+)`)
)

func brewPrefix() string {
	out, err := exec.Command("brew", "--prefix").Output()
	if err != nil {
		return "/usr/local"
	}
	return strings.TrimSpace(string(out))
}

func phpBinFor(version string) string {
	return brewPrefix() + "/opt/php@" + version + "/bin/php"
}

func phpCandidates(version string) []string {
	return []string{
		phpBinFor(version),
		"/usr/local/bin/php" + version,
	}
}

func runOut(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func phpVersion(bin string) string {
	out, err := runOut(bin, "-v")
	if err != nil {
		return ""
	}
	if m := versionRe.FindStringSubmatch(out); len(m) > 1 {
		return m[1]
	}
	return ""
}

func extVersion(bin, ext string) string {
	out, err := runOut(bin, "--ri", ext)
	if err != nil {
		return ""
	}
	if m := extVerRe.FindStringSubmatch(out); len(m) > 1 {
		return m[1]
	}
	return ""
}

func extLoaded(bin, ext string) bool {
	_, err := runOut(bin, "--ri", ext)
	return err == nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
	out, err := runOut("brew", "services", "list")
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
