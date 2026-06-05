package index

import (
	indexquery "code-index-plugin/internal/index/query"
	indexservice "code-index-plugin/internal/index/service"
)

type BuildRequest = indexservice.BuildRequest

type BuildResult = indexservice.BuildResult

type RefreshRequest = indexservice.RefreshRequest

type RefreshResult = indexservice.RefreshResult

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
