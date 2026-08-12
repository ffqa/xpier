package ca

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"xpier/internal/nginx"
	"xpier/internal/store"
)

func CaPaths() (string, string) {
	return filepath.Join(store.XpierHome(), "ca", "xpier-ca.pem"),
		filepath.Join(store.XpierHome(), "ca", "xpier-ca-key.pem")
}

// EnsureCA generates a local CA (if missing) so site certs can be trusted by
// the system keychain with no browser warnings.
func EnsureCA() error {
	cert, key := CaPaths()
	if store.FileExists(cert) && store.FileExists(key) {
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

// TrustCA installs the CA into the system keychain (requires root).
func TrustCA() error {
	cert, _ := CaPaths()
	if err := EnsureCA(); err != nil {
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

// EnsureWildcardCertSignedByCA replaces the self-signed wildcard cert with one
// signed by the xpier CA once the CA exists.
func EnsureWildcardCertSignedByCA(tld string) error {
	cert, key := nginx.CertPaths(tld)
	if store.FileExists(cert) && store.FileExists(key) {
		// Already present; leave it (first secure run replaces it).
		return nil
	}
	caCert, caKey := CaPaths()
	if !store.FileExists(caCert) {
		return nil // CA not created yet; keep self-signed
	}
	csr := filepath.Join(store.XpierHome(), "ca", "wildcard."+tld+".csr")
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
	ext := filepath.Join(store.XpierHome(), "ca", "wildcard."+tld+".ext")
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

// nginx.DomainCertPaths returns cert/key paths for a specific domain.
// EnsureDomainCert signs a cert for a specific domain (SAN: domain + wildcard
// of its base, e.g. img.test28.test -> *.test28.test) with the xpier CA.
func EnsureDomainCert(domain string) error {
	cert, key := nginx.DomainCertPaths(domain)
	if store.FileExists(cert) && store.FileExists(key) {
		return nil
	}
	caCert, caKey := CaPaths()
	if !store.FileExists(caCert) {
		return fmt.Errorf("xpier CA missing; run `sudo xpier secure` first")
	}
	// base = domain without its first label (img.test28.test -> test28.test);
	// a single-label domain has no wildcard base.
	labels := strings.SplitN(domain, ".", 2)
	wild := ""
	if len(labels) == 2 {
		wild = "*." + labels[1]
	}
	csr := filepath.Join(store.XpierHome(), "ca", "domain-"+domain+".csr")
	ext := filepath.Join(store.XpierHome(), "ca", "domain-"+domain+".ext")
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

// CmdSecure trusts the xpier CA and signs certs. With a domain argument it
// signs a per-domain cert (e.g. `xpier secure img.test28`), otherwise it
// signs the *.test wildcard cert.
func CmdSecure(args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("usage: sudo xpier secure [domain]    trust CA + sign certs")
		return nil
	}
	if os.Geteuid() != 0 {
		return fmt.Errorf("`xpier secure` must run as root: use `sudo xpier secure`")
	}
	if err := TrustCA(); err != nil {
		return err
	}
	if len(args) > 0 {
		arg := strings.TrimPrefix(args[0], ".")
		if !store.SafeSiteNameRe.MatchString(arg) {
			return fmt.Errorf("invalid domain %q", arg)
		}
		// Resolve the full domain: a linked site name becomes
		// <name>.<tld>; otherwise append the TLD unless already full.
		domain := arg
		linked := false
		if sites, err := store.LoadSites(); err == nil {
			if _, ok := sites.Sites[arg]; ok {
				domain = store.SiteDomain(sites, arg)
				linked = true
				// Re-enable https for a site that was `unsecure`d.
				if site := sites.Sites[arg]; site.Secure != nil && !*site.Secure {
					site.Secure = nil
					sites.Sites[arg] = site
					sites.Save()
				}
			} else if !strings.HasSuffix(arg, "."+sites.TLD) {
				domain = arg + "." + sites.TLD
			}
		}
		_ = linked
		if err := EnsureDomainCert(domain); err != nil {
			return err
		}
		if base := strings.SplitN(domain, ".", 2); len(base) == 2 {
			fmt.Printf("xpier CA trusted; cert for %s signed (SAN: %s + *.%s)\n", domain, domain, base[1])
		} else {
			fmt.Printf("xpier CA trusted; cert for %s signed\n", domain)
		}
	} else {
		// Drop the existing self-signed cert so it gets re-signed by the CA.
		tld := nginx.CurrentTLD()
		cert, _ := nginx.CertPaths(tld)
		os.Remove(cert)
		if err := EnsureWildcardCertSignedByCA(tld); err != nil {
			return err
		}
		fmt.Printf("xpier CA trusted; *.%s certs are now signed by it (browsers will not warn)\n", tld)
	}
	if err := nginx.NginxReload(); err != nil {
		return fmt.Errorf("nginx reload after secure failed: %w (run `sudo xpier install` first?)", err)
	}
	// Regenerate site configs so they pick up the freshly signed cert.
	if sites, err := store.LoadSites(); err == nil {
		if err := nginx.WriteAllSiteConfigs(sites); err != nil {
			return err
		}
		if err := nginx.NginxReload(); err != nil {
			return fmt.Errorf("nginx reload after secure failed: %w", err)
		}
	}
	return nil
}

func CmdSecured(args []string) error {
	sites, err := store.LoadSites()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(sites.Sites))
	for name := range sites.Sites {
		names = append(names, name)
	}
	var httpOnly []string
	for _, name := range names {
		site := sites.Sites[name]
		if site.Secure != nil && !*site.Secure {
			httpOnly = append(httpOnly, name)
			continue
		}
		fmt.Printf("  %-30s https://%s\n", store.SiteDomain(sites, name), store.SiteDomain(sites, name))
	}
	if len(httpOnly) > 0 {
		fmt.Println("http-only (xpier unsecure):")
		for _, name := range httpOnly {
			fmt.Printf("  %-30s http://%s\n", store.SiteDomain(sites, name), store.SiteDomain(sites, name))
		}
	}
	fmt.Println("sites listen on 80+443 by default (self-signed wildcard); `sudo xpier secure` signs a trusted cert to silence browser warnings")
	return nil
}
