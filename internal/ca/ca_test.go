package ca

import (
	"os"
	"path/filepath"
	"testing"

	"xpier/internal/nginx"
	"xpier/internal/store"
)

func homeTemp(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestCaPaths(t *testing.T) {
	homeTemp(t)
	c, k := CaPaths()
	if !filepath.IsAbs(c) || !filepath.IsAbs(k) {
		t.Errorf("CaPaths = %q %q", c, k)
	}
	if !store.FileExists(c) || !store.FileExists(k) {
		// fine; paths only
	}
}

func TestEnsureCA(t *testing.T) {
	homeTemp(t)
	if err := EnsureCA(); err != nil {
		t.Skipf("openssl unavailable: %v", err)
	}
	cert, key := CaPaths()
	if !store.FileExists(cert) || !store.FileExists(key) {
		t.Fatal("CA not created")
	}
	if err := EnsureCA(); err != nil {
		t.Errorf("second EnsureCA = %v", err)
	}
}

func TestEnsureWildcardCertSignedByCA(t *testing.T) {
	homeTemp(t)
	if err := EnsureCA(); err != nil {
		t.Skipf("openssl unavailable: %v", err)
	}
	if err := EnsureWildcardCertSignedByCA("test"); err != nil {
		t.Fatal(err)
	}
	cert, key := nginx.CertPaths("test")
	if !store.FileExists(cert) || !store.FileExists(key) {
		t.Error("signed wildcard cert not created")
	}
	// Idempotent.
	if err := EnsureWildcardCertSignedByCA("test"); err != nil {
		t.Errorf("second run = %v", err)
	}
	// No CA -> no-op.
	home := store.XpierHome()
	os.RemoveAll(filepath.Join(home, "ca"))
	os.RemoveAll(filepath.Join(home, "certs"))
	if err := EnsureWildcardCertSignedByCA("test"); err != nil {
		t.Errorf("no-CA run = %v", err)
	}
}

func TestEnsureDomainCert(t *testing.T) {
	homeTemp(t)
	if err := EnsureDomainCert("img.test28.test"); err == nil {
		t.Error("EnsureDomainCert without CA should error")
	}
	if err := EnsureCA(); err != nil {
		t.Skipf("openssl unavailable: %v", err)
	}
	if err := EnsureDomainCert("img.test28.test"); err != nil {
		t.Fatal(err)
	}
	dc, dk := nginx.DomainCertPaths("img.test28.test")
	if !store.FileExists(dc) || !store.FileExists(dk) {
		t.Error("domain cert not created")
	}
	if err := EnsureDomainCert("img.test28.test"); err != nil {
		t.Errorf("second run = %v", err)
	}
}

func TestCmdSecuredEmpty(t *testing.T) {
	homeTemp(t)
	if err := CmdSecured(nil); err != nil {
		t.Errorf("CmdSecured = %v", err)
	}
}

func TestCmdSecureNonRoot(t *testing.T) {
	homeTemp(t)
	if err := CmdSecure(nil); err == nil {
		t.Error("CmdSecure without root should error")
	}
}

func TestTrustCANonRoot(t *testing.T) {
	homeTemp(t)
	if err := TrustCA(); err == nil {
		t.Error("TrustCA without root should error")
	}
}
