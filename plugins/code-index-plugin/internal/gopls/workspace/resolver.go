package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrNoWorkspace = errors.New("未找到 Go workspace")

type Kind string

const (
	GoWork Kind = "go.work"
	GoMod  Kind = "go.mod"
)

type Workspace struct {
	Root     string `json:"root"`
	Kind     Kind   `json:"kind"`
	Manifest string `json:"manifest"`
}

type Resolver struct{}

func NewResolver() *Resolver {
	return &Resolver{}
}

// Discover resolves an enclosing workspace first, then scans child directories
// for independent workspaces when path is a multi-repository parent directory.
func (r *Resolver) Discover(path string) ([]Workspace, error) {
	dir, err := Normalize(path)
	if err != nil {
		return nil, err
	}

	if ws, ok := nearest(dir); ok {
		return []Workspace{ws}, nil
	}

	var workspaces []Workspace
	err = filepath.WalkDir(dir, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if current != dir && ignoredDirectory(entry.Name()) {
			return filepath.SkipDir
		}

		if ws, ok := markerAt(current); ok {
			workspaces = append(workspaces, ws)
			if current != dir {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("扫描 Go workspace 失败: %w", err)
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].Root < workspaces[j].Root
	})
	return workspaces, nil
}

func (r *Resolver) Resolve(path string) (Workspace, error) {
	dir, err := Normalize(path)
	if err != nil {
		return Workspace{}, err
	}
	if ws, ok := nearest(dir); ok {
		return ws, nil
	}
	return Workspace{}, fmt.Errorf("%w: %s", ErrNoWorkspace, dir)
}

func (r *Resolver) ResolveFile(file string) (Workspace, error) {
	if !filepath.IsAbs(file) {
		return Workspace{}, fmt.Errorf("Go 文件路径必须是绝对路径: %s", file)
	}
	info, err := os.Stat(file)
	if err != nil {
		return Workspace{}, fmt.Errorf("访问 Go 文件失败: %w", err)
	}
	if info.IsDir() {
		return Workspace{}, fmt.Errorf("Go 文件路径指向目录: %s", file)
	}
	return r.Resolve(filepath.Dir(file))
}

func Normalize(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("路径不能为空")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("解析绝对路径失败: %w", err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("解析路径失败: %w", err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", fmt.Errorf("访问路径失败: %w", err)
	}
	if !info.IsDir() {
		real = filepath.Dir(real)
	}
	return filepath.Clean(real), nil
}

func Contains(root, target string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func nearest(dir string) (Workspace, bool) {
	var nearestModule Workspace
	for current := dir; ; current = filepath.Dir(current) {
		if manifest := regularFile(filepath.Join(current, "go.work")); manifest != "" {
			return Workspace{Root: current, Kind: GoWork, Manifest: manifest}, true
		}
		if nearestModule.Root == "" {
			if manifest := regularFile(filepath.Join(current, "go.mod")); manifest != "" {
				nearestModule = Workspace{Root: current, Kind: GoMod, Manifest: manifest}
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nearestModule, nearestModule.Root != ""
		}
	}
}

func markerAt(dir string) (Workspace, bool) {
	for _, candidate := range []struct {
		name string
		kind Kind
	}{
		{name: "go.work", kind: GoWork},
		{name: "go.mod", kind: GoMod},
	} {
		manifest := regularFile(filepath.Join(dir, candidate.name))
		if manifest != "" {
			return Workspace{Root: dir, Kind: candidate.kind, Manifest: manifest}, true
		}
	}
	return Workspace{}, false
}

func regularFile(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.Mode().IsRegular() {
		return path
	}
	return ""
}

func ignoredDirectory(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", ".claude", ".idea", ".vscode", "node_modules", "vendor":
		return true
	default:
		return false
	}
}
