package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code-index-plugin/internal/index/extractor"
)

func TestScanSkipsIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "service", "user.go"), "package service")
	mustWriteFile(t, filepath.Join(root, "node_modules", "pkg", "index.js"), "export {}")
	mustWriteFile(t, filepath.Join(root, ".git", "objects", "head"), "ref")
	mustWriteFile(t, filepath.Join(root, "vendor", "dep", "dep.go"), "package dep")
	mustWriteFile(t, filepath.Join(root, "README.md"), "# ignored by extension filter")

	got, err := New(DefaultOptions()).Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 file candidate, got %d", len(got))
	}
	if got[0].Path != "service/user.go" {
		t.Fatalf("expected relative path service/user.go, got %q", got[0].Path)
	}
	if got[0].AbsPath != filepath.Join(root, "service", "user.go") {
		t.Fatalf("expected absolute path %q, got %q", filepath.Join(root, "service", "user.go"), got[0].AbsPath)
	}
	if got[0].Language != "go" {
		t.Fatalf("expected language go, got %q", got[0].Language)
	}
}

func TestBuildFileRecordIncludesSummaryAndRoleTag(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "controller", "payment.go")
	mustWriteFile(t, path, "package controller\n\n// Payment callback handler\nfunc Handle() {}")

	candidates, err := New(DefaultOptions()).Scan(root)
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}

	rec, err := extractor.BuildFileRecord(root, candidates[0])
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	if rec.Path != "controller/payment.go" {
		t.Fatalf("expected record path controller/payment.go, got %q", rec.Path)
	}
	if rec.Language != "go" {
		t.Fatalf("expected language go, got %q", rec.Language)
	}
	if rec.Hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !contains(rec.RoleTags, "controller") {
		t.Fatalf("expected role tags to contain controller, got %v", rec.RoleTags)
	}
	if !containsSubstring(rec.Summary, "Payment") {
		t.Fatalf("expected summary to contain Payment, got %q", rec.Summary)
	}
	if !contains(rec.Keywords, "payment") {
		t.Fatalf("expected keywords to contain payment, got %v", rec.Keywords)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsSubstring(value, substr string) bool {
	return strings.Contains(value, substr)
}
