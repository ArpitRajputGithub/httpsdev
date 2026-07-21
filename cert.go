package main

import (
	"os"
	"time"
)

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
