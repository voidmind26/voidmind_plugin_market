package model

import "time"

// FileCandidate 表示扫描阶段发现、可进入索引流程的文件。
type FileCandidate struct {
	Path     string
	AbsPath  string
	Language string
	Size     int64
	ModTime  time.Time
}

// FileRecord 表示文件级索引记录。
type FileRecord struct {
	Path        string   `json:"path"`
	Language    string   `json:"language"`
	Size        int64    `json:"size"`
	MTime       int64    `json:"mtime"`
	Hash        string   `json:"hash"`
	Summary     string   `json:"summary,omitempty"`
	RoleTags    []string `json:"role_tags,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	ModuleHints []string `json:"module_hints,omitempty"`
	Imports     []string `json:"imports,omitempty"`
}

// SymbolRecord 表示从源码中抽取出的结构化符号。
type SymbolRecord struct {
	Path       string   `json:"path"`
	Language   string   `json:"language"`
	SymbolName string   `json:"symbol_name"`
	SymbolType string   `json:"symbol_type"`
	Receiver   string   `json:"receiver,omitempty"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	Summary    string   `json:"summary,omitempty"`
	Keywords   []string `json:"keywords,omitempty"`
}

// ChunkRecord 表示从源码或文档中按启发式提取出的内容块。
type ChunkRecord struct {
	Path      string   `json:"path"`
	Language  string   `json:"language"`
	ChunkType string   `json:"chunk_type"`
	Title     string   `json:"title"`
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
	Summary   string   `json:"summary,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

type ManifestFileState struct {
	Path     string `json:"path"`
	Hash     string `json:"hash"`
	Size     int64  `json:"size"`
	MTime    int64  `json:"mtime"`
	Language string `json:"language,omitempty"`
}

type Manifest struct {
	Files       map[string]ManifestFileState `json:"files,omitempty"`
	DataFiles   map[string]string            `json:"data_files,omitempty"`
	GeneratedAt int64                        `json:"generated_at,omitempty"`
}

type ManifestDiff struct {
	Added     []string `json:"added,omitempty"`
	Changed   []string `json:"changed,omitempty"`
	Deleted   []string `json:"deleted,omitempty"`
	Unchanged []string `json:"unchanged,omitempty"`
}

// ProjectIndex 为索引持久化载体，包含文件级记录与最小深索引结果。
type ProjectIndex struct {
	Manifest Manifest        `json:"manifest"`
	Files    []FileRecord    `json:"files,omitempty"`
	Symbols  []SymbolRecord  `json:"symbols,omitempty"`
	Chunks   []ChunkRecord   `json:"chunks,omitempty"`
}
