package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	indexconfig "code-index-plugin/internal/index/config"
	"code-index-plugin/internal/index/model"
)

type Options = indexconfig.Options

type Scanner struct {
	opts Options
}

func DefaultOptions() Options {
	return indexconfig.DefaultOptions()
}

func New(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

func (s *Scanner) Scan(root string) ([]model.FileCandidate, error) {
	root = indexconfig.NormalizeProjectRoot(root)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var out []model.FileCandidate
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if s.shouldSkipDir(d.Name(), path, root) {
				return filepath.SkipDir
			}
			return nil
		}

		candidate, ok, err := s.buildCandidate(root, path)
		if err != nil {
			return err
		}
		if ok {
			out = append(out, candidate)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func (s *Scanner) shouldSkipDir(name, path, root string) bool {
	if path == root {
		return false
	}
	_, ok := s.opts.IgnoredDirs[name]
	return ok
}

func (s *Scanner) buildCandidate(root, path string) (model.FileCandidate, bool, error) {
	ext := strings.ToLower(filepath.Ext(path))
	language, ok := s.opts.AllowedExtensions[ext]
	if !ok {
		return model.FileCandidate{}, false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return model.FileCandidate{}, false, err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return model.FileCandidate{}, false, err
	}

	return model.FileCandidate{
		Path:     filepath.ToSlash(rel),
		AbsPath:  path,
		Language: language,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
	}, true, nil
}
