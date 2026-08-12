package xpier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func caPaths() (string, string) {
	return filepath.Join(xpierHome(), "ca", "xpier-ca.pem"),
		filepath.Join(xpierHome(), "ca", "xpier-ca-key.pem")
}

// ensureCA generates a local CA (if missing) so site certs can be trusted by
// the system keychain with no browser warnings.
func ensureCA() error {
	cert, key := caPaths()
	if fileExists(cert) && fileExists(key) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(cert), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-keyout", key, "-out", cert, "-days", "3650",
		"-subj", "/CN=Herdy Local CA",
		"-addext", "basicConstraints=critical,CA:TRUE",
		"-addext", "keyUsage=critical,keyCertSign,cRLSign")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("openssl CA: %v: %s", err, out)
	}
	return nil
}

// trustCA installs the CA into the system keychain (requires root).
func trustCA() error {
	cert, _ := caPaths()
	if err := ensureCA(); err != nil {
		return err
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("trusting the CA requires root: run `sudo xpier secure` or `sudo xpier install`")
	}
	out, err := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot",
		"-k", "/Library/Keychains/System.keychain", cert).CombinedOutput()
	if err != nil {
		return fmt.Errorf("security add-trusted-cert: %v: %s", err, out)
	}
	return nil
}

// ensureWildcardCertSignedByCA replaces the self-signed wildcard cert with one
// signed by the xpier CA once the CA exists.
func ensureWildcardCertSignedByCA(tld string) error {
	cert, key := certPaths(tld)
	if fileExists(cert) && fileExists(key) {
		// Already present; leave it (first secure run replaces it).
		return nil
	}
	caCert, caKey := caPaths()
	if !fileExists(caCert) {
		return nil // CA not created yet; keep self-signed
	}
	csr := filepath.Join(xpierHome(), "ca", "wildcard."+tld+".csr")
	if err := os.MkdirAll(filepath.Dir(cert), 0o755); err != nil {
		return err
	}
	run := func(args ...string) error {
		out, err := exec.Command("openssl", args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("openssl %s: %v: %s", args[0], err, out)
		}
		return nil
	}
	// Generate key + CSR for *.tld.
	if err := run("req", "-newkey", "rsa:2048", "-nodes",
		"-keyout", key, "-out", csr,
		"-subj", "/CN=*."+tld,
		"-addext", "subjectAltName=DNS:*."+tld+",DNS:"+tld+",DNS:localhost"); err != nil {
		return err
	}
	// Sign with the CA; SANs go through an extfile.
	ext := filepath.Join(xpierHome(), "ca", "wildcard."+tld+".ext")
	if err := os.WriteFile(ext, []byte("subjectAltName=DNS:*."+tld+",DNS:"+tld+",DNS:localhost"), 0o644); err != nil {
		return err
	}
	if err := run("x509", "-req", "-in", csr, "-CA", caCert, "-CAkey", caKey,
		"-CAcreateserial", "-out", cert, "-days", "3650", "-extfile", ext); err != nil {
		return err
	}
	os.Remove(csr)
	os.Remove(ext)
	return nil
}

// domainCertPaths returns cert/key paths for a specific domain.
func domainCertPaths(domain string) (string, string) {
	return filepath.Join(xpierHome(), "certs", domain+".pem"),
		filepath.Join(xpierHome(), "certs", domain+"-key.pem")
}

// ensureDomainCert signs a cert for a specific domain (SAN: domain + wildcard
// of its base, e.g. img.test28.test -> *.test28.test) with the xpier CA.
func ensureDomainCert(domain string) error {
	cert, key := domainCertPaths(domain)
	if fileExists(cert) && fileExists(key) {
		return nil
	}
	caCert, caKey := caPaths()
	if !fileExists(caCert) {
		return fmt.Errorf("xpier CA missing; run `sudo xpier secure` first")
	}
	// base = domain without its first label (img.test28.test -> test28.test);
	// a single-label domain has no wildcard base.
	labels := strings.SplitN(domain, ".", 2)
	wild := ""
	if len(labels) == 2 {
		wild = "*." + labels[1]
	}
	csr := filepath.Join(xpierHome(), "ca", "domain-"+domain+".csr")
	ext := filepath.Join(xpierHome(), "ca", "domain-"+domain+".ext")
	if err := os.MkdirAll(filepath.Dir(cert), 0o755); err != nil {
		return err
	}
	run := func(args ...string) error {
		out, err := exec.Command("openssl", args...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("openssl %s: %v: %s", args[0], err, out)
		}
		return nil
	}
	san := "subjectAltName=DNS:" + domain + ",DNS:localhost"
	if wild != "" {
		san += ",DNS:" + wild
	}
	if err := os.WriteFile(ext, []byte(san), 0o644); err != nil {
		return err
	}
	if err := run("req", "-newkey", "rsa:2048", "-nodes",
		"-keyout", key, "-out", csr, "-subj", "/CN="+domain); err != nil {
		return err
	}
	if err := run("x509", "-req", "-in", csr, "-CA", caCert, "-CAkey", caKey,
		"-CAcreateserial", "-out", cert, "-days", "3650", "-extfile", ext); err != nil {
		return err
	}
	os.Remove(csr)
	os.Remove(ext)
	return nil
}

// cmdSecure trusts the xpier CA and signs certs. With a domain argument it
// signs a per-domain cert (e.g. `xpier secure img.test28`), otherwise it
// signs the *.test wildcard cert.
func cmdSecure(args []string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("`xpier secure` must run as root: use `sudo xpier secure`")
	}
	if err := trustCA(); err != nil {
		return err
	}
	if len(args) > 0 {
		arg := strings.TrimPrefix(args[0], ".")
		if !safeSiteNameRe.MatchString(arg) {
			return fmt.Errorf("invalid domain %q", arg)
		}
		// Resolve the full domain: a linked site name becomes
		// <name>.<tld>; otherwise append the TLD unless already full.
		domain := arg
		if sites, err := loadSites(); err == nil {
			if _, ok := sites.Sites[arg]; ok {
				domain = siteDomain(sites, arg)
			} else if !strings.HasSuffix(arg, "."+sites.TLD) {
				domain = arg + "." + sites.TLD
			}
		}
		if err := ensureDomainCert(domain); err != nil {
			return err
		}
		if base := strings.SplitN(domain, ".", 2); len(base) == 2 {
			fmt.Printf("xpier CA trusted; cert for %s signed (SAN: %s + *.%s)\n", domain, domain, base[1])
		} else {
			fmt.Printf("xpier CA trusted; cert for %s signed\n", domain)
		}
	} else {
		// Drop the existing self-signed cert so it gets re-signed by the CA.
		cert, _ := certPaths("test")
		os.Remove(cert)
		if err := ensureWildcardCertSignedByCA("test"); err != nil {
			return err
		}
		fmt.Println("xpier CA trusted; *.test certs are now signed by it (browsers will not warn)")
	}
	if err := nginxReload(); err != nil {
		fmt.Printf("[warn] nginx reload failed: %v\n", err)
	}
	// Regenerate site configs so they pick up the freshly signed cert.
	if sites, err := loadSites(); err == nil {
		writeAllSiteConfigs(sites)
		nginxReload()
	}
	return nil
}

func cmdSecured(args []string) error {
	sites, err := loadSites()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(sites.Sites))
	for name := range sites.Sites {
		names = append(names, name)
	}
	for _, name := range names {
		fmt.Printf("  %-30s https://%s\n", siteDomain(sites, name), siteDomain(sites, name))
	}
	return nil
}

var _ = strings.TrimSpace
