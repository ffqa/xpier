// Package store is the data layer for xpier: shared base helpers, types, and
// all persistence (sites, proxies, app states, manifests, locks).
package store

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/yaml.v3"
)

const (
	ManifestName = "xpier.yaml"
	LockName     = "xpier.lock"
)

// --- types ---

type Manifest struct {
	PHP        string            `yaml:"php,omitempty"`
	Runtime    string            `yaml:"runtime,omitempty"`
	Extensions map[string]string `yaml:"extensions,omitempty"`
	Services   []string          `yaml:"services,omitempty"`
	Apps       map[string]App    `yaml:"apps,omitempty"`
}

type App struct {
	Dir        string            `yaml:"dir"`
	Cmd        string            `yaml:"cmd"`
	Port       string            `yaml:"port,omitempty"`
	Ports      []string          `yaml:"ports,omitempty"`
	Domain     string            `yaml:"domain,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	Node       string            `yaml:"node,omitempty"`
	PHP        string            `yaml:"php,omitempty"`
	Extensions []string          `yaml:"extensions,omitempty"`
}

type AppConfig struct {
	Namespace string         `yaml:"namespace,omitempty"`
	Apps      map[string]App `yaml:"apps"`
}

type Sites struct {
	TLD    string          `yaml:"tld"`
	Parked []string        `yaml:"parked,omitempty"`
	Sites  map[string]Site `yaml:"sites"`
}

type Site struct {
	Path   string `json:"path"`
	PHP    string `json:"php,omitempty"`
	Node   string `json:"node,omitempty"`
	Driver string `json:"driver"`
}

type Lock struct {
	SchemaVersion int           `yaml:"schema_version"`
	GeneratedAt   string        `yaml:"generated_at"`
	PHP           PhpLock       `yaml:"php"`
	Extensions    []ExtLock     `yaml:"extensions"`
	Services      []ServiceLock `yaml:"services"`
}

type PhpLock struct {
	Version string `yaml:"version"`
	Path    string `yaml:"path"`
}

type ExtLock struct {
	Name       string `yaml:"name"`
	Constraint string `yaml:"constraint"`
	Installed  string `yaml:"installed"`
	Loaded     bool   `yaml:"loaded"`
}

type ServiceLock struct {
	Name    string `yaml:"name"`
	Running bool   `yaml:"running"`
}

type AppState struct {
	Name   string   `json:"name"`
	PID    int      `json:"pid"`
	Cmd    string   `json:"cmd,omitempty"`
	Log    string   `json:"log"`
	Port   string   `json:"port"`
	Ports  []string `json:"ports,omitempty"`
	Domain string   `json:"domain"`
}

// --- base helpers ---

func XpierHome() string {
	// When run via `sudo xpier install`, keep using the real user's home.
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		if u, err := user.Lookup(sudoUser); err == nil {
			return filepath.Join(u.HomeDir, ".xpier")
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return filepath.Join(home, ".xpier")
}

func SlugFor(dir string) string {
	abs, _ := filepath.Abs(dir)
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:4])
}

func SlugName(dir string) string {
	abs, _ := filepath.Abs(dir)
	return filepath.Base(abs) + "_" + SlugFor(abs)
}

func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

// ProcAlive reports whether pid is alive AND its command line contains marker
// (when marker is non-empty). Kills must be guarded with a marker so a PID
// recycled after a reboot can never take down an unrelated process.
func ProcAlive(pid int, marker string) bool {
	if !PidAlive(pid) {
		return false
	}
	if marker == "" {
		return true
	}
	out, err := RunOut("ps", "-o", "command=", "-p", strconv.Itoa(pid))
	if err != nil {
		return false
	}
	return strings.Contains(out, marker)
}

func KillGroup(pid int, sig syscall.Signal) error {
	err := syscall.Kill(-pid, sig)
	if err == syscall.ESRCH {
		return syscall.Kill(pid, sig)
	}
	return err
}

func RunOut(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func BrewPrefix() string {
	out, err := exec.Command("brew", "--prefix").Output()
	if err != nil {
		return "/usr/local"
	}
	return strings.TrimSpace(string(out))
}

func PortBusy(port string) (bool, error) {
	out, err := exec.Command("lsof", "-ti", "tcp:"+port, "-sTCP:LISTEN").Output()
	if err != nil && strings.TrimSpace(string(out)) == "" {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func ConfirmYesNo(prompt string) (bool, error) {
	fmt.Printf("%s [y/N] ", prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	}
	return false, nil
}

func YAMLUnmarshal(data []byte, v any) error { return yaml.Unmarshal(data, v) }

// --- manifest / lock paths ---

func ProjectPaths(dir string) (string, string) {
	base := filepath.Join(XpierHome(), "projects", SlugName(dir))
	return filepath.Join(base, ManifestName), filepath.Join(base, LockName)
}

func ResolvePaths(dir string) (string, string) {
	localManifest := filepath.Join(dir, ManifestName)
	if _, err := os.Stat(localManifest); err == nil {
		return localManifest, filepath.Join(dir, LockName)
	}
	return ProjectPaths(dir)
}

func EnsureProjectDir(dir string) error {
	base, _ := ProjectPaths(dir)
	return os.MkdirAll(filepath.Dir(base), 0o755)
}

func DefaultManifest() *Manifest {
	return &Manifest{Runtime: "fpm"}
}

func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o644)
}

func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func LoadLock(path string) (*Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Lock
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (l *Lock) Save(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o644)
}

// --- sites registry ---

func SitesPath() string { return filepath.Join(XpierHome(), "sites.json") }

func DefaultSites() *Sites {
	return &Sites{TLD: "test", Sites: map[string]Site{}}
}

func LoadSites() (*Sites, error) {
	data, err := os.ReadFile(SitesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSites(), nil
		}
		return nil, err
	}
	var s Sites
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.TLD == "" {
		s.TLD = "test"
	}
	if s.Sites == nil {
		s.Sites = map[string]Site{}
	}
	return &s, nil
}

func (s *Sites) Save() error {
	if err := os.MkdirAll(filepath.Dir(SitesPath()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(SitesPath(), data, 0o644)
}

// --- proxies registry ---

func ProxiesPath() string { return filepath.Join(XpierHome(), "proxies.json") }

func LoadProxies() (map[string]string, error) {
	data, err := os.ReadFile(ProxiesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]string{}
	}
	return m, nil
}

func SaveProxies(m map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(ProxiesPath()), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(ProxiesPath(), data, 0o644)
}

// --- app state ---

func AppStatePath(ns, name string) string {
	return filepath.Join(XpierHome(), "apps", ns, name+".json")
}

func AppLogPath(ns, name string) string {
	return filepath.Join(XpierHome(), "apps", ns, "logs", "dev-"+name+".log")
}

func LoadAppState(ns, name string) (*AppState, error) {
	data, err := os.ReadFile(AppStatePath(ns, name))
	if err != nil {
		return nil, err
	}
	var s AppState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveAppState(s *AppState, ns string) error {
	if err := os.MkdirAll(filepath.Dir(AppStatePath(ns, s.Name)), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return writeFileAtomic(AppStatePath(ns, s.Name), data, 0o644)
}

// EnsureBrewPackage prompts to install a missing brew package and installs it
// on confirmation. Returns nil when the binary is already present.
func EnsureBrewPackage(bin, formula, display string) error {
	if FileExists(bin) {
		return nil
	}
	ok, err := ConfirmYesNo(fmt.Sprintf("%s 未安装（brew install %s），是否现在安装？", display, formula))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%s not installed; run `brew install %s` and retry", display, formula)
	}
	fmt.Printf("installing %s...\n", formula)
	if out, err := exec.Command("brew", "install", formula).CombinedOutput(); err != nil {
		return fmt.Errorf("brew install %s: %v: %s", formula, err, out)
	}
	return nil
}

func CurrentUser() (*user.User, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && os.Geteuid() == 0 {
		return user.Lookup(sudoUser)
	}
	return user.Current()
}

func SiteDomain(s *Sites, name string) string { return name + "." + s.TLD }

func SiteRoot(site Site) string {
	switch site.Driver {
	case "laravel":
		return filepath.Join(site.Path, "public")
	case "spa":
		return filepath.Join(site.Path, "dist")
	}
	return site.Path
}

var SafeSiteNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

var SafePhpRe = regexp.MustCompile(`^\d+\.\d+$`)

func RunOutErr(name string, args ...string) error {
	_, err := RunOut(name, args...)
	return err
}

func UDPBusy(port string) (bool, error) {
	out, err := RunOut("lsof", "-ti", "udp:"+port)
	if err != nil && out == "" {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func DnsmasqConfPath() string {
	return filepath.Join(XpierHome(), "dnsmasq", "dnsmasq.conf")
}

func WriteDnsmasqConfig(tld string) error {
	conf := fmt.Sprintf(`port=53
listen-address=127.0.0.1
bind-interfaces
no-resolv
address=/.%s/127.0.0.1
`, tld)
	if err := os.MkdirAll(filepath.Dir(DnsmasqConfPath()), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(DnsmasqConfPath(), []byte(conf), 0o644)
}

// paintWord maps a status word to its ANSI color: up=green, down=red,
// none/no=*yellow. Plain strings pass through untouched.
func paintWord(s string) string {
	switch {
	case s == "up" || strings.HasPrefix(s, "up"):
		return "\x1b[32m" + s + "\x1b[0m"
	case s == "down":
		return "\x1b[31m" + s + "\x1b[0m"
	case s == "none" || strings.HasPrefix(s, "no "):
		return "\x1b[33m" + s + "\x1b[0m"
	}
	return s
}

// colorEnabled reports whether ANSI color output is appropriate: NO_COLOR is
// unset and stdout is a terminal (not a pipe).
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// Paint colors the status words used across xpier output. Returns s unchanged
// when colors are disabled (piped output, NO_COLOR).
func Paint(s string) string {
	if !colorEnabled() {
		return s
	}
	return paintWord(s)
}

// writeFileAtomic writes data to path via a temp file + rename so a crash
// mid-write never leaves a truncated registry behind.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// UpDown formats a boolean as the up/down string used by status tables.
func UpDown(up bool) string {
	if up {
		return Paint("up")
	}
	return Paint("down")
}
