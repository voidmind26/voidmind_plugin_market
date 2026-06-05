package config

import "strings"

type Options struct {
	IgnoredDirs       map[string]struct{}
	AllowedExtensions map[string]string
}

func DefaultOptions() Options {
	return Options{
		IgnoredDirs: map[string]struct{}{
			".claude":      {},
			".git":         {},
			".idea":        {},
			".vscode":      {},
			"bin":          {},
			"dist":         {},
			"build":        {},
			"coverage":     {},
			"node_modules": {},
			"vendor":       {},
			"tmp":          {},
			"temp":         {},
			"__pycache__":  {},
			".turbo":       {},
			".next":        {},
			".nuxt":        {},
			"target":       {},
		},
		AllowedExtensions: map[string]string{
			".go":   "go",
			".js":   "javascript",
			".jsx":  "javascript",
			".ts":   "typescript",
			".tsx":  "typescript",
			".py":   "python",
			".java": "java",
			".rb":   "ruby",
			".php":  "php",
			".rs":   "rust",
			".sh":   "shell",
			".sql":  "sql",
			".json": "json",
			".yaml": "yaml",
			".yml":  "yaml",
			".xml":  "xml",
			".html": "html",
			".css":  "css",
			".scss": "scss",
		},
	}
}

func NormalizeProjectRoot(root string) string {
	return strings.TrimSpace(root)
}
