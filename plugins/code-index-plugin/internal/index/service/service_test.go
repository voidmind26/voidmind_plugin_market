package service

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"code-index-plugin/internal/index/scanner"
	"code-index-plugin/internal/index/storage"
)

func TestBuildCreatesIndexFiles(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "service", "user.go"), "package service\n\nfunc Handle() {}")
	mustWriteFile(t, filepath.Join(root, "web", "overview.ts"), "export function overview() {}\n")

	svc := New(storage.New(), scanner.New(DefaultOptions()))
	result, err := svc.Build(context.Background(), BuildRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	if result.FileCount != 2 {
		t.Fatalf("expected file count 2, got %d", result.FileCount)
	}
	if result.SymbolCount < 1 {
		t.Fatalf("expected at least 1 symbol, got %d", result.SymbolCount)
	}
	if result.ChunkCount != 2 {
		t.Fatalf("expected 2 chunks so go and ts files both contribute chunks, got %d", result.ChunkCount)
	}
	if result.IndexDir != filepath.Join(root, ".claude", "code-index") {
		t.Fatalf("unexpected index dir %q", result.IndexDir)
	}

	assertFileExists(t, filepath.Join(root, ".claude", "code-index", "manifest.json"))
	assertFileExists(t, filepath.Join(root, ".claude", "code-index", "files.jsonl"))
	assertFileExists(t, filepath.Join(root, ".claude", "code-index", "symbols.jsonl"))
	chunksPath := filepath.Join(root, ".claude", "code-index", "chunks.jsonl")
	assertFileExists(t, chunksPath)
	assertFileContains(t, chunksPath, `"path":"service/user.go"`)
}

func TestRefreshReturnsErrorWhenManifestExistsButChunksMissing(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "service", "user.go"), "package service\n\nfunc Handle() {}")

	svc := New(storage.New(), scanner.New(DefaultOptions()))
	result, err := svc.Build(context.Background(), BuildRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if err := os.Remove(filepath.Join(result.IndexDir, "chunks.jsonl")); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = svc.Refresh(context.Background(), RefreshRequest{ProjectRoot: root})
	if err == nil {
		t.Fatal("expected Refresh to fail when manifest exists but chunks.jsonl is missing")
	}
	if !strings.Contains(err.Error(), "chunks.jsonl") {
		t.Fatalf("expected error to mention chunks.jsonl, got %v", err)
	}
}

func TestRefreshReturnsAddedChangedDeletedAndUnchangedCounts(t *testing.T) {
	root := t.TempDir()
	alpha := filepath.Join(root, "service", "alpha.go")
	beta := filepath.Join(root, "service", "beta.go")
	keep := filepath.Join(root, "web", "keep.ts")
	mustWriteFile(t, alpha, "package service\n\nfunc Alpha() {}")
	mustWriteFile(t, beta, "package service\n\nfunc Beta() {}")
	mustWriteFile(t, keep, "export const keep = true\n")

	svc := New(storage.New(), scanner.New(DefaultOptions()))
	if _, err := svc.Build(context.Background(), BuildRequest{ProjectRoot: root}); err != nil {
		t.Fatalf("initial Build returned error: %v", err)
	}

	mustWriteFile(t, alpha, "package service\n\nfunc Alpha() {\n\tprintln(\"changed\")\n}")
	if err := os.Remove(beta); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	mustWriteFile(t, filepath.Join(root, "service", "gamma.go"), "package service\n\nfunc Gamma() {}")
	mustWriteFile(t, filepath.Join(root, "web", "delta.ts"), "export const delta = true\n")

	result, err := svc.Refresh(context.Background(), RefreshRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	if result.FileCount != 4 {
		t.Fatalf("expected refreshed file count 4, got %d", result.FileCount)
	}
	if result.AddedCount != 2 {
		t.Fatalf("expected added count 2, got %d", result.AddedCount)
	}
	if result.ChangedCount != 1 {
		t.Fatalf("expected changed count 1, got %d", result.ChangedCount)
	}
	if result.DeletedCount != 1 {
		t.Fatalf("expected deleted count 1, got %d", result.DeletedCount)
	}
	if result.UnchangedCount != 1 {
		t.Fatalf("expected unchanged count 1, got %d", result.UnchangedCount)
	}

	if !reflect.DeepEqual([]string{"service/gamma.go", "web/delta.ts"}, result.Added) {
		t.Fatalf("unexpected added paths %v", result.Added)
	}
	if !reflect.DeepEqual([]string{"service/alpha.go"}, result.Changed) {
		t.Fatalf("unexpected changed paths %v", result.Changed)
	}
	if !reflect.DeepEqual([]string{"service/beta.go"}, result.Deleted) {
		t.Fatalf("unexpected deleted paths %v", result.Deleted)
	}
	if state, ok := result.Manifest.Files["web/delta.ts"]; !ok || state.Path != "web/delta.ts" {
		t.Fatalf("expected manifest to contain web/delta.ts, got %v", result.Manifest.Files)
	}
}

func TestSearchReturnsRankedStructuredResultsFromStoredIndex(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "service", "payment.go"), "package service\n\nfunc HandlePaymentCallback() error {\n\treturn nil\n}\n")
	mustWriteFile(t, filepath.Join(root, "docs", "payment.md"), "# payment callback\n")

	svc := New(storage.New(), scanner.New(DefaultOptions()))
	if _, err := svc.Build(context.Background(), BuildRequest{ProjectRoot: root}); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	result, err := svc.Search(context.Background(), SearchRequest{ProjectRoot: root, Query: "payment callback", Limit: 3})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if result.ProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, result.ProjectRoot)
	}
	if result.Query != "payment callback" {
		t.Fatalf("expected query to echo input, got %q", result.Query)
	}
	if result.ResultCount == 0 || len(result.Results) == 0 {
		t.Fatal("expected structured search results")
	}
	if result.Results[0].Kind == "" || result.Results[0].Path == "" || result.Results[0].Score <= 0 {
		t.Fatalf("expected first result to contain structured hit data, got %+v", result.Results[0])
	}
}

func TestStatusReturnsIndexMetadataWhenIndexExists(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "service", "payment.go"), "package service\n\nfunc HandlePaymentCallback() error {\n\treturn nil\n}\n")
	mustWriteFile(t, filepath.Join(root, "web", "payment.ts"), "export function renderPayment() {}\n")

	svc := New(storage.New(), scanner.New(DefaultOptions()))
	buildResult, err := svc.Build(context.Background(), BuildRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	status, err := svc.Status(context.Background(), StatusRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if !status.Ready {
		t.Fatal("expected status to report ready=true")
	}
	if status.ProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, status.ProjectRoot)
	}
	if status.IndexDir != buildResult.IndexDir {
		t.Fatalf("expected index dir %q, got %q", buildResult.IndexDir, status.IndexDir)
	}
	if status.FileCount != buildResult.FileCount || status.SymbolCount != buildResult.SymbolCount || status.ChunkCount != buildResult.ChunkCount {
		t.Fatalf("expected counts to match build result, got status=%+v build=%+v", status, buildResult)
	}
	if status.GeneratedAt == 0 {
		t.Fatal("expected generatedAt to be populated")
	}
}

func TestStatusReturnsNotReadyWhenIndexDoesNotExist(t *testing.T) {
	root := t.TempDir()
	svc := New(storage.New(), scanner.New(DefaultOptions()))

	status, err := svc.Status(context.Background(), StatusRequest{ProjectRoot: root})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if status.Ready {
		t.Fatal("expected status to report ready=false when index is missing")
	}
	if status.ProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, status.ProjectRoot)
	}
	if status.IndexDir == "" {
		t.Fatal("expected index dir to be populated even when index is missing")
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

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %q to exist: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, needle string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), needle) {
		t.Fatalf("expected file %q to contain %q, got %s", path, needle, string(data))
	}
}
