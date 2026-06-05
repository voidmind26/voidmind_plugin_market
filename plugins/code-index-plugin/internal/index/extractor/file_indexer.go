package extractor

import (
	"crypto/sha256"
	"encoding/hex"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	indexconfig "code-index-plugin/internal/index/config"
	"code-index-plugin/internal/index/model"
)

var wordPattern = regexp.MustCompile(`[A-Za-z][A-Za-z0-9_]{2,}`)

const maxKeywordCount = 256

func BuildFileRecord(root string, candidate model.FileCandidate) (model.FileRecord, error) {
	content, err := os.ReadFile(candidate.AbsPath)
	if err != nil {
		return model.FileRecord{}, err
	}
	return BuildFileRecordFromContent(root, candidate, content)
}

func BuildFileRecordFromContent(root string, candidate model.FileCandidate, content []byte) (model.FileRecord, error) {
	root = indexconfig.NormalizeProjectRoot(root)
	if root == "" {
		return model.FileRecord{}, os.ErrInvalid
	}
	if candidate.AbsPath == "" {
		return model.FileRecord{}, os.ErrInvalid
	}

	rel := filepath.ToSlash(candidate.Path)
	if rel == "" {
		var err error
		rel, err = filepath.Rel(root, candidate.AbsPath)
		if err != nil {
			return model.FileRecord{}, err
		}
		rel = filepath.ToSlash(rel)
	}
	language := candidate.Language
	if language == "" {
		language = detectLanguage(candidate.AbsPath)
	}
	roleTags := inferRoleTags(rel)
	keywords := collectKeywords(rel, content)
	summary := buildFileSummary(rel, content)
	mtime := candidate.ModTime.Unix()
	if candidate.ModTime.IsZero() {
		info, err := os.Stat(candidate.AbsPath)
		if err != nil {
			return model.FileRecord{}, err
		}
		candidate.Size = info.Size()
		mtime = info.ModTime().Unix()
	}

	return model.FileRecord{
		Path:        rel,
		Language:    language,
		Size:        candidate.Size,
		MTime:       mtime,
		Hash:        sha256Hex(content),
		ModuleHints: moduleHints(rel),
		Imports:     extractImports(language, candidate.AbsPath, content),
		Keywords:    keywords,
		Summary:     summary,
		RoleTags:    roleTags,
	}, nil
}

func detectLanguage(path string) string {
	return indexconfig.DefaultOptions().AllowedExtensions[strings.ToLower(filepath.Ext(path))]
}

func inferRoleTags(rel string) []string {
	parts := strings.Split(rel, "/")
	known := map[string]struct{}{
		"api": {}, "cmd": {}, "config": {}, "controller": {}, "handler": {}, "model": {},
		"repo": {}, "repository": {}, "route": {}, "router": {}, "service": {}, "store": {},
		"test": {}, "util": {}, "utils": {}, "view": {},
	}

	seen := map[string]struct{}{}
	var tags []string
	for _, part := range parts[:max(0, len(parts)-1)] {
		part = strings.ToLower(strings.TrimSpace(part))
		if _, ok := known[part]; ok {
			if _, exists := seen[part]; !exists {
				seen[part] = struct{}{}
				tags = append(tags, part)
			}
		}
	}
	return tags
}

func collectKeywords(rel string, content []byte) []string {
	_ = rel
	text := string(content)
	seen := make(map[string]struct{}, maxKeywordCount)
	keywords := make([]string, 0, maxKeywordCount)
	for offset := 0; offset < len(text) && len(keywords) < maxKeywordCount; {
		loc := wordPattern.FindStringIndex(text[offset:])
		if loc == nil {
			break
		}
		start := offset + loc[0]
		end := offset + loc[1]
		for _, term := range splitTerms(text[start:end]) {
			if len(keywords) >= maxKeywordCount {
				break
			}
			if _, exists := seen[term]; exists {
				continue
			}
			seen[term] = struct{}{}
			keywords = append(keywords, term)
		}
		offset = end
	}
	sort.Strings(keywords)
	return keywords
}

func buildFileSummary(rel string, content []byte) string {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
		}
		if strings.HasPrefix(trimmed, "#") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		}
	}
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	base = strings.ReplaceAll(base, "_", " ")
	if base == "" {
		return rel
	}
	return strings.Title(base)
}

func moduleHints(rel string) []string {
	parts := strings.Split(rel, "/")
	if len(parts) <= 1 {
		return nil
	}
	return []string{parts[0]}
}

func extractImports(language, path string, content []byte) []string {
	if language != "go" {
		return nil
	}

	parsed, err := parser.ParseFile(token.NewFileSet(), path, content, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		if spec == nil || spec.Path == nil {
			continue
		}
		unquoted, err := strconv.Unquote(spec.Path.Value)
		if err != nil || unquoted == "" {
			continue
		}
		imports = append(imports, unquoted)
	}
	return imports
}

func sha256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
