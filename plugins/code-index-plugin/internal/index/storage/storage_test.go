package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"code-index-plugin/internal/index/scanner"
	servicepkg "code-index-plugin/internal/index/service"
	"code-index-plugin/internal/index/storage"
)

func TestLoadProjectIndexSupportsLargeJSONLRecords(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "docs", "big.ts"), buildLargeFileContent(12000))

	svc := servicepkg.New(storage.New(), scanner.New(servicepkg.DefaultOptions()))
	buildResult, err := svc.Build(context.Background(), servicepkg.BuildRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if buildResult.FileCount != 1 {
		t.Fatalf("expected file count 1, got %d", buildResult.FileCount)
	}

	payload, err := storage.New().LoadProjectIndex(root)
	if err != nil {
		t.Fatalf("LoadProjectIndex returned error: %v", err)
	}
	if len(payload.Files) != 1 {
		t.Fatalf("expected 1 file record, got %d", len(payload.Files))
	}

	filesPath := filepath.Join(root, ".claude", "code-index", "files.jsonl")
	original, err := os.ReadFile(filesPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	mustWriteFile(t, filesPath, string(original)+string(original))

	_, err = storage.New().LoadProjectIndex(root)
	if err == nil {
		t.Fatal("expected LoadProjectIndex to fail after tampering files.jsonl")
	}
	if !strings.Contains(err.Error(), "files.jsonl") || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("expected digest mismatch error for files.jsonl, got %v", err)
	}
}

func buildLargeFileContent(count int) string {
	parts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		parts = append(parts, "keywordtoken_"+strconv.Itoa(i)+strings.Repeat("x", 12))
	}
	return "export const huge = '" + strings.Join(parts, " ") + "'\n"
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
