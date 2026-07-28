package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsIndependentChildModules(t *testing.T) {
	root := t.TempDir()
	moduleA := writeMarker(t, root, "repo-a", "go.mod")
	moduleB := writeMarker(t, root, "group/repo-b", "go.mod")
	writeMarker(t, moduleB, "vendor/nested", "go.mod")
	moduleA = normalized(t, moduleA)
	moduleB = normalized(t, moduleB)

	got, err := NewResolver().Discover(root)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Discover() returned %d workspaces, want 2: %+v", len(got), got)
	}
	roots := map[string]bool{got[0].Root: true, got[1].Root: true}
	if !roots[moduleA] || !roots[moduleB] {
		t.Fatalf("Discover() roots = %v, want %s and %s", roots, moduleA, moduleB)
	}
}

func TestDiscoverPrefersEnclosingGoWork(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.work"), "go 1.24\n")
	module := writeMarker(t, root, "module", "go.mod")
	root = normalized(t, root)

	got, err := NewResolver().Discover(module)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(got) != 1 || got[0].Root != root || got[0].Kind != GoWork {
		t.Fatalf("Discover() = %+v, want enclosing go.work %s", got, root)
	}
}

func TestResolveFileUsesNearestModule(t *testing.T) {
	root := t.TempDir()
	outer := writeMarker(t, root, "outer", "go.mod")
	inner := writeMarker(t, outer, "inner", "go.mod")
	file := filepath.Join(inner, "main.go")
	writeFile(t, file, "package main\n")
	inner = normalized(t, inner)

	got, err := NewResolver().ResolveFile(file)
	if err != nil {
		t.Fatalf("ResolveFile() error = %v", err)
	}
	if got.Root != inner {
		t.Fatalf("ResolveFile() root = %s, want %s", got.Root, inner)
	}
}

func normalized(t *testing.T, path string) string {
	t.Helper()
	got, err := Normalize(path)
	if err != nil {
		t.Fatalf("Normalize(%s): %v", path, err)
	}
	return got
}

func TestContainsRejectsSiblingPrefix(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "tmp", "repo")
	if Contains(root, filepath.Join(string(filepath.Separator), "tmp", "repo-other", "main.go")) {
		t.Fatal("Contains() accepted a sibling with the same path prefix")
	}
}

func writeMarker(t *testing.T, base, relative, marker string) string {
	t.Helper()
	dir := filepath.Join(base, relative)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	writeFile(t, filepath.Join(dir, marker), "module example.com/test\n")
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
