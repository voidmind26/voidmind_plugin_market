package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"code-index-plugin/internal/index/extractor"
	"code-index-plugin/internal/index/manifest"
	"code-index-plugin/internal/index/model"
	indexquery "code-index-plugin/internal/index/query"
	"code-index-plugin/internal/index/scanner"
	"code-index-plugin/internal/index/storage"
)

type BuildRequest struct {
	ProjectRoot    string   `json:"project_root,omitempty"`
	DeepIndexPaths []string `json:"deep_index_paths,omitempty"`
}

type BuildResult struct {
	ProjectRoot    string         `json:"project_root,omitempty"`
	IndexDir       string         `json:"index_dir,omitempty"`
	DeepIndexPaths []string       `json:"deep_index_paths,omitempty"`
	FileCount      int            `json:"file_count,omitempty"`
	SymbolCount    int            `json:"symbol_count,omitempty"`
	ChunkCount     int            `json:"chunk_count,omitempty"`
	Manifest       model.Manifest `json:"manifest"`
}

type RefreshRequest struct {
	ProjectRoot    string   `json:"project_root,omitempty"`
	DeepIndexPaths []string `json:"deep_index_paths,omitempty"`
}

type RefreshResult struct {
	ProjectRoot    string         `json:"project_root,omitempty"`
	IndexDir       string         `json:"index_dir,omitempty"`
	DeepIndexPaths []string       `json:"deep_index_paths,omitempty"`
	FileCount      int            `json:"file_count,omitempty"`
	SymbolCount    int            `json:"symbol_count,omitempty"`
	ChunkCount     int            `json:"chunk_count,omitempty"`
	AddedCount     int            `json:"added_count,omitempty"`
	ChangedCount   int            `json:"changed_count,omitempty"`
	DeletedCount   int            `json:"deleted_count,omitempty"`
	UnchangedCount int            `json:"unchanged_count,omitempty"`
	Added          []string       `json:"added,omitempty"`
	Changed        []string       `json:"changed,omitempty"`
	Deleted        []string       `json:"deleted,omitempty"`
	Unchanged      []string       `json:"unchanged,omitempty"`
	Manifest       model.Manifest `json:"manifest"`
}

type SearchRequest struct {
	Query          string `json:"query"`
	ProjectRoot    string `json:"project_root,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	PreferDeepHits bool   `json:"prefer_deep_hits,omitempty"`
	PathPrefix     string `json:"path_prefix,omitempty"`
}

type SearchHit = indexquery.SearchHit

type SearchResult struct {
	Query       string      `json:"query"`
	ProjectRoot string      `json:"project_root,omitempty"`
	Limit       int         `json:"limit,omitempty"`
	ResultCount int         `json:"result_count"`
	Results     []SearchHit `json:"results"`
}

type StatusRequest struct {
	ProjectRoot string `json:"project_root,omitempty"`
}

type StatusResult struct {
	ProjectRoot string `json:"project_root,omitempty"`
	Ready       bool   `json:"ready"`
	IndexDir    string `json:"index_dir,omitempty"`
	FileCount   int    `json:"file_count,omitempty"`
	SymbolCount int    `json:"symbol_count,omitempty"`
	ChunkCount  int    `json:"chunk_count,omitempty"`
	GeneratedAt int64  `json:"generated_at,omitempty"`
}

type Service struct {
	storage *storage.Storage
	scanner *scanner.Scanner
}

func New(store *storage.Storage, scan *scanner.Scanner) *Service {
	if store == nil {
		store = storage.New()
	}
	if scan == nil {
		scan = scanner.New(DefaultOptions())
	}
	return &Service{storage: store, scanner: scan}
}

func (s *Service) Build(ctx context.Context, req BuildRequest) (BuildResult, error) {
	root, err := resolveProjectRoot(req.ProjectRoot)
	if err != nil {
		return BuildResult{}, err
	}
	payload, err := s.buildIndexPayload(ctx, root, req.DeepIndexPaths)
	if err != nil {
		return BuildResult{}, err
	}
	if err := s.storage.SaveProjectIndex(root, payload); err != nil {
		return BuildResult{}, err
	}
	indexDir, err := s.storage.IndexDir(root)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{
		ProjectRoot:    root,
		IndexDir:       indexDir,
		DeepIndexPaths: normalizeDeepIndexPaths(req.DeepIndexPaths),
		FileCount:      len(payload.Files),
		SymbolCount:    len(payload.Symbols),
		ChunkCount:     len(payload.Chunks),
		Manifest:       payload.Manifest,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (RefreshResult, error) {
	root, err := resolveProjectRoot(req.ProjectRoot)
	if err != nil {
		return RefreshResult{}, err
	}
	previous, err := s.storage.LoadProjectIndex(root)
	if err != nil {
		return RefreshResult{}, err
	}
	payload, err := s.buildIndexPayload(ctx, root, req.DeepIndexPaths)
	if err != nil {
		return RefreshResult{}, err
	}
	diff := manifest.Diff(previous.Manifest, payload.Manifest)
	if err := s.storage.SaveProjectIndex(root, payload); err != nil {
		return RefreshResult{}, err
	}
	indexDir, err := s.storage.IndexDir(root)
	if err != nil {
		return RefreshResult{}, err
	}
	return RefreshResult{
		ProjectRoot:    root,
		IndexDir:       indexDir,
		DeepIndexPaths: normalizeDeepIndexPaths(req.DeepIndexPaths),
		FileCount:      len(payload.Files),
		SymbolCount:    len(payload.Symbols),
		ChunkCount:     len(payload.Chunks),
		AddedCount:     len(diff.Added),
		ChangedCount:   len(diff.Changed),
		DeletedCount:   len(diff.Deleted),
		UnchangedCount: len(diff.Unchanged),
		Added:          diff.Added,
		Changed:        diff.Changed,
		Deleted:        diff.Deleted,
		Unchanged:      diff.Unchanged,
		Manifest:       payload.Manifest,
	}, nil
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (SearchResult, error) {
	root, err := resolveProjectRoot(req.ProjectRoot)
	if err != nil {
		return SearchResult{}, err
	}
	payload, err := s.storage.LoadProjectIndex(root)
	if err != nil {
		return SearchResult{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	resp := indexquery.Search(payload, indexquery.SearchRequest{
		Query:          req.Query,
		Limit:          limit,
		PreferDeepHits: req.PreferDeepHits,
		PathPrefix:     req.PathPrefix,
	})
	return SearchResult{
		Query:       req.Query,
		ProjectRoot: root,
		Limit:       limit,
		ResultCount: resp.ResultCount,
		Results:     resp.Results,
	}, nil
}

func (s *Service) Status(ctx context.Context, req StatusRequest) (StatusResult, error) {
	_ = ctx
	root, err := resolveProjectRoot(req.ProjectRoot)
	if err != nil {
		return StatusResult{}, err
	}
	indexDir, err := s.storage.IndexDir(root)
	if err != nil {
		return StatusResult{}, err
	}
	payload, err := s.storage.LoadProjectIndex(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return StatusResult{
				ProjectRoot: root,
				Ready:       false,
				IndexDir:    indexDir,
			}, nil
		}
		return StatusResult{}, err
	}
	return StatusResult{
		ProjectRoot: root,
		Ready:       true,
		IndexDir:    indexDir,
		FileCount:   len(payload.Files),
		SymbolCount: len(payload.Symbols),
		ChunkCount:  len(payload.Chunks),
		GeneratedAt: payload.Manifest.GeneratedAt,
	}, nil
}

func (s *Service) buildIndexPayload(ctx context.Context, root string, deepIndexPaths []string) (model.ProjectIndex, error) {
	candidates, err := s.scanner.Scan(root)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	prefixes := normalizeDeepIndexPaths(deepIndexPaths)
	files := make([]model.FileRecord, 0, len(candidates))
	var symbols []model.SymbolRecord
	var chunks []model.ChunkRecord
	for _, candidate := range candidates {
		select {
		case <-ctx.Done():
			return model.ProjectIndex{}, ctx.Err()
		default:
		}
		content, err := os.ReadFile(candidate.AbsPath)
		if err != nil {
			return model.ProjectIndex{}, fmt.Errorf("read candidate %s: %w", candidate.Path, err)
		}
		record, err := extractor.BuildFileRecordFromContent(root, candidate, content)
		if err != nil {
			return model.ProjectIndex{}, fmt.Errorf("build file record %s: %w", candidate.Path, err)
		}
		files = append(files, record)
		if record.Language == "go" && shouldDeepIndex(record.Path, prefixes) {
			recordSymbols, err := extractor.ExtractGoSymbols(record.Path, content)
			if err != nil {
				return model.ProjectIndex{}, fmt.Errorf("extract go symbols %s: %w", record.Path, err)
			}
			symbols = append(symbols, recordSymbols...)
			chunks = append(chunks, extractor.ExtractHeuristicChunks(record.Path, record.Language, content)...)
			continue
		}
		chunks = append(chunks, extractor.ExtractHeuristicChunks(record.Path, record.Language, content)...)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].Path == symbols[j].Path {
			if symbols[i].StartLine == symbols[j].StartLine {
				return symbols[i].SymbolName < symbols[j].SymbolName
			}
			return symbols[i].StartLine < symbols[j].StartLine
		}
		return symbols[i].Path < symbols[j].Path
	})
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].Path == chunks[j].Path {
			if chunks[i].StartLine == chunks[j].StartLine {
				return chunks[i].Title < chunks[j].Title
			}
			return chunks[i].StartLine < chunks[j].StartLine
		}
		return chunks[i].Path < chunks[j].Path
	})
	manifestData := manifest.Build(files)
	return model.ProjectIndex{
		Manifest: manifestData,
		Files:    files,
		Symbols:  symbols,
		Chunks:   chunks,
	}, nil
}

func shouldDeepIndex(path string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	path = filepath.ToSlash(path)
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func normalizeDeepIndexPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		normalized := filepath.ToSlash(strings.TrimSpace(path))
		normalized = strings.Trim(normalized, "/")
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func resolveProjectRoot(root string) (string, error) {
	root = NormalizeProjectRoot(root)
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = cwd
	}
	return filepath.Clean(root), nil
}
