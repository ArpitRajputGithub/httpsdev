package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldRegenerate_MissingFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "localhost.pem")
	keyPath := filepath.Join(dir, "localhost.key")

	if !shouldRegenerate(certPath, keyPath, 30*24*time.Hour) {
		t.Fatal("expected regenerate=true when files missing")
	}
}

func TestShouldRegenerate_FreshFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "localhost.pem")
	keyPath := filepath.Join(dir, "localhost.key")
	writeFile(t, certPath, "cert")
	writeFile(t, keyPath, "key")

	if shouldRegenerate(certPath, keyPath, 30*24*time.Hour) {
		t.Fatal("expected regenerate=false when files fresh")
	}
}

func TestShouldRegenerate_OldFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "localhost.pem")
	keyPath := filepath.Join(dir, "localhost.key")
	writeFile(t, certPath, "cert")
	writeFile(t, keyPath, "key")
	old := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(certPath, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(keyPath, old, old); err != nil {
		t.Fatal(err)
	}

	if !shouldRegenerate(certPath, keyPath, 30*24*time.Hour) {
		t.Fatal("expected regenerate=true when files older than max age")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
