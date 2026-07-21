package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// certCacheDir returns the directory where cached certs live.
// Follows XDG Base Directory spec.
func certCacheDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "httpsdev", "certs")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "httpsdev", "certs")
}

// shouldRegenerate returns true if either cert file is missing or older than maxAge.
func shouldRegenerate(certPath, keyPath string, maxAge time.Duration) bool {
	certInfo, err := os.Stat(certPath)
	if err != nil {
		return true
	}
	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		return true
	}
	cutoff := time.Now().Add(-maxAge)
	return certInfo.ModTime().Before(cutoff) || keyInfo.ModTime().Before(cutoff)
}

// ensureCert returns paths to a valid cached cert+key, regenerating via mkcert if needed.
// extraHosts is added to the mkcert SAN list in addition to localhost, 127.0.0.1, ::1.
func ensureCert(extraHosts []string) (certPath, keyPath string, err error) {
	dir := certCacheDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", fmt.Errorf("create cert cache dir: %w", err)
	}
	certPath = filepath.Join(dir, "localhost.pem")
	keyPath = filepath.Join(dir, "localhost.key")

	if !shouldRegenerate(certPath, keyPath, 30*24*time.Hour) {
		return certPath, keyPath, nil
	}

	if _, err := exec.LookPath("mkcert"); err != nil {
		return "", "", errors.New(
			"mkcert not found on PATH.\n" +
				"install it first:\n" +
				"  brew install mkcert && mkcert -install",
		)
	}

	hosts := append([]string{"localhost", "127.0.0.1", "::1"}, extraHosts...)
	args := append([]string{"-cert-file", certPath, "-key-file", keyPath}, hosts...)
	cmd := exec.Command("mkcert", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("mkcert failed: %w\n%s", err, string(out))
	}
	return certPath, keyPath, nil
}
