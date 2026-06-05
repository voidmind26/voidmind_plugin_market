package storage

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	indexconfig "code-index-plugin/internal/index/config"
	"code-index-plugin/internal/index/model"
)

const (
	indexDirName     = ".claude/code-index"
	manifestFileName = "manifest.json"
	filesFileName    = "files.jsonl"
	symbolsFileName  = "symbols.jsonl"
	chunksFileName   = "chunks.jsonl"
)

type Storage struct{}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) SaveProjectIndex(root string, payload model.ProjectIndex) error {
	indexDir, err := s.IndexDir(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(indexDir, 0o755); err != nil {
		return err
	}
	filesPath := filepath.Join(indexDir, filesFileName)
	symbolsPath := filepath.Join(indexDir, symbolsFileName)
	chunksPath := filepath.Join(indexDir, chunksFileName)
	if err := writeJSONL(filesPath, payload.Files); err != nil {
		return err
	}
	if err := writeJSONL(symbolsPath, payload.Symbols); err != nil {
		return err
	}
	if err := writeJSONL(chunksPath, payload.Chunks); err != nil {
		return err
	}
	if payload.Manifest.DataFiles == nil {
		payload.Manifest.DataFiles = map[string]string{}
	}
	payload.Manifest.DataFiles[filesFileName], err = fileSHA256(filesPath)
	if err != nil {
		return err
	}
	payload.Manifest.DataFiles[symbolsFileName], err = fileSHA256(symbolsPath)
	if err != nil {
		return err
	}
	payload.Manifest.DataFiles[chunksFileName], err = fileSHA256(chunksPath)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(indexDir, manifestFileName), payload.Manifest); err != nil {
		return err
	}
	return nil
}

func (s *Storage) LoadProjectIndex(root string) (model.ProjectIndex, error) {
	indexDir, err := s.IndexDir(root)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	manifest, err := s.LoadManifest(root)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	filesPath := filepath.Join(indexDir, filesFileName)
	symbolsPath := filepath.Join(indexDir, symbolsFileName)
	chunksPath := filepath.Join(indexDir, chunksFileName)
	if err := ensureExists(filesPath); err != nil {
		return model.ProjectIndex{}, err
	}
	if err := ensureExists(symbolsPath); err != nil {
		return model.ProjectIndex{}, err
	}
	if err := ensureExists(chunksPath); err != nil {
		return model.ProjectIndex{}, err
	}
	if err := verifyDigest(filesPath, manifest.DataFiles[filesFileName]); err != nil {
		return model.ProjectIndex{}, err
	}
	if err := verifyDigest(symbolsPath, manifest.DataFiles[symbolsFileName]); err != nil {
		return model.ProjectIndex{}, err
	}
	if err := verifyDigest(chunksPath, manifest.DataFiles[chunksFileName]); err != nil {
		return model.ProjectIndex{}, err
	}
	files, err := loadJSONL[model.FileRecord](filesPath)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	symbols, err := loadJSONL[model.SymbolRecord](symbolsPath)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	chunks, err := loadJSONL[model.ChunkRecord](chunksPath)
	if err != nil {
		return model.ProjectIndex{}, err
	}
	return model.ProjectIndex{
		Manifest: manifest,
		Files:    files,
		Symbols:  symbols,
		Chunks:   chunks,
	}, nil
}

func (s *Storage) LoadManifest(root string) (model.Manifest, error) {
	indexDir, err := s.IndexDir(root)
	if err != nil {
		return model.Manifest{}, err
	}
	var manifest model.Manifest
	if err := readJSON(filepath.Join(indexDir, manifestFileName), &manifest); err != nil {
		return model.Manifest{}, err
	}
	if manifest.Files == nil {
		manifest.Files = map[string]model.ManifestFileState{}
	}
	if manifest.DataFiles == nil {
		manifest.DataFiles = map[string]string{}
	}
	return manifest, nil
}

func (s *Storage) IndexDir(root string) (string, error) {
	root = indexconfig.NormalizeProjectRoot(root)
	if root == "" {
		return "", os.ErrInvalid
	}
	return filepath.Join(root, filepath.FromSlash(indexDirName)), nil
}

func readJSON(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func writeJSON(path string, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'))
}

func writeJSONL[T any](path string, rows []T) error {
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(tmpPath)
	}()

	writer := bufio.NewWriter(file)
	for _, row := range rows {
		data, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		if err := writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func writeAtomic(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(tmpPath)
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func ensureExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("index data file missing: %s", filepath.Base(path))
		}
		return err
	}
	return nil
}

func verifyDigest(path, expected string) error {
	if expected == "" {
		return fmt.Errorf("index digest missing: %s", filepath.Base(path))
	}
	actual, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("index digest mismatch: %s", filepath.Base(path))
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func loadJSONL[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var out []T
	decoder := json.NewDecoder(bufio.NewReader(file))
	for {
		var row T
		if err := decoder.Decode(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}
