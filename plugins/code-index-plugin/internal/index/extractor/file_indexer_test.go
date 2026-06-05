package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code-index-plugin/internal/index/model"
)

func TestBuildFileRecordExtractsSingleLineGoImport(t *testing.T) {
	root := t.TempDir()
	absPath := filepath.Join(root, "service", "payment.go")
	mustWriteFile(t, absPath, "package service\n\nimport \"fmt\"\n\nfunc Handle() { _ = fmt.Sprintf(\"ok\") }")

	rec, err := BuildFileRecord(root, model.FileCandidate{
		Path:     "service/payment.go",
		AbsPath:  absPath,
		Language: "go",
		ModTime:  time.Now(),
	})
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	if len(rec.Imports) != 1 || rec.Imports[0] != "fmt" {
		t.Fatalf("expected imports [fmt], got %v", rec.Imports)
	}
}

func TestBuildFileRecordExtractsImportBlockWithoutEmptyEntries(t *testing.T) {
	root := t.TempDir()
	absPath := filepath.Join(root, "service", "payment.go")
	mustWriteFile(t, absPath, "package service\n\nimport (\n\t\"fmt\"\n\talias \"strings\"\n\t. \"math\"\n\t_ \"net/http/pprof\"\n)\n\nfunc Handle() { _, _ = fmt.Println, alias.TrimSpace; _ = Abs(-1) }")

	rec, err := BuildFileRecord(root, model.FileCandidate{
		Path:     "service/payment.go",
		AbsPath:  absPath,
		Language: "go",
		ModTime:  time.Now(),
	})
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	if len(rec.Imports) != 4 {
		t.Fatalf("expected 4 imports, got %d (%v)", len(rec.Imports), rec.Imports)
	}
	assertContains(t, rec.Imports, "fmt")
	assertContains(t, rec.Imports, "strings")
	assertContains(t, rec.Imports, "math")
	assertContains(t, rec.Imports, "net/http/pprof")
	for _, item := range rec.Imports {
		if item == "" {
			t.Fatalf("expected no empty import entries, got %v", rec.Imports)
		}
	}
}

func TestBuildFileRecordReusesCandidatePathAndLanguage(t *testing.T) {
	root := t.TempDir()
	absPath := filepath.Join(root, "nested", "handler.custom")
	mustWriteFile(t, absPath, "# custom file")

	rec, err := BuildFileRecord(root, model.FileCandidate{
		Path:     "virtual/handler.custom",
		AbsPath:  absPath,
		Language: "customlang",
		ModTime:  time.Now(),
	})
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	if rec.Path != "virtual/handler.custom" {
		t.Fatalf("expected record path to reuse candidate path, got %q", rec.Path)
	}
	if rec.Language != "customlang" {
		t.Fatalf("expected record language to reuse candidate language, got %q", rec.Language)
	}
}

func TestBuildFileRecordKeywordsAreBoundedAndNormalized(t *testing.T) {
	root := t.TempDir()
	absPath := filepath.Join(root, "docs", "big.ts")
	content := ""
	for i := 0; i < 600; i++ {
		content += fmt.Sprintf("Token_%03d TOKEN_%03d token_%03d ", i, i, i)
	}
	mustWriteFile(t, absPath, content)

	rec, err := BuildFileRecord(root, model.FileCandidate{
		Path:     "docs/big.ts",
		AbsPath:  absPath,
		Language: "typescript",
		ModTime:  time.Now(),
	})
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	if len(rec.Keywords) != maxKeywordCount {
		t.Fatalf("expected %d keywords, got %d", maxKeywordCount, len(rec.Keywords))
	}
	seen := make(map[string]struct{}, len(rec.Keywords))
	for _, keyword := range rec.Keywords {
		if keyword != strings.ToLower(keyword) {
			t.Fatalf("expected lowercase keyword, got %q", keyword)
		}
		if _, ok := seen[keyword]; ok {
			t.Fatalf("expected deduplicated keywords, found duplicate %q", keyword)
		}
		seen[keyword] = struct{}{}
	}
}

func TestBuildFileRecordKeywordsComeFromContentNotPath(t *testing.T) {
	root := t.TempDir()
	absPath := filepath.Join(root, "controllers", "login.go")
	mustWriteFile(t, absPath, "package helpers\n\nfunc GetLoginUserInfo() {}")

	rec, err := BuildFileRecord(root, model.FileCandidate{
		Path:     "controllers/login.go",
		AbsPath:  absPath,
		Language: "go",
		ModTime:  time.Now(),
	})
	if err != nil {
		t.Fatalf("BuildFileRecord returned error: %v", err)
	}
	for _, keyword := range rec.Keywords {
		if keyword == "controllers" {
			t.Fatalf("expected path-only term to be excluded from keywords, got %v", rec.Keywords)
		}
	}
	assertContains(t, rec.Keywords, "getloginuserinfo")
	assertContains(t, rec.Keywords, "helpers")
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

func assertContains(t *testing.T, items []string, target string) {
	t.Helper()
	for _, item := range items {
		if item == target {
			return
		}
	}
	t.Fatalf("expected %q in %v", target, items)
}
