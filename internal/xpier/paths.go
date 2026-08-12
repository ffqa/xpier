package xpier

import (
	"os"
	"path/filepath"
)

const (
	manifestName = "xpier.yaml"
	lockName     = "xpier.lock"
)

// slugName is a readable per-project key: "basename_hash8".
func slugName(dir string) string {
	abs, _ := filepath.Abs(dir)
	return filepath.Base(abs) + "_" + slugFor(abs)
}

// projectPaths is the default storage under ~/.xpier/projects/<name>/
// so project repos stay clean.
func projectPaths(dir string) (string, string) {
	base := filepath.Join(xpierHome(), "projects", slugName(dir))
	return filepath.Join(base, manifestName), filepath.Join(base, lockName)
}

// resolvePaths picks the manifest/lock location for a project: an in-repo
// xpier.yaml (opt-in via `xpier init .`) wins, otherwise ~/.xpier is used.
func resolvePaths(dir string) (string, string) {
	localManifest := filepath.Join(dir, manifestName)
	if _, err := os.Stat(localManifest); err == nil {
		return localManifest, filepath.Join(dir, lockName)
	}
	return projectPaths(dir)
}

func ensureProjectDir(dir string) error {
	base, _ := projectPaths(dir)
	return os.MkdirAll(filepath.Dir(base), 0o755)
}
