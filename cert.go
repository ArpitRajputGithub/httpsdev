package main

import (
	"os"
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
