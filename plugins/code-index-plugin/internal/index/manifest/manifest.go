package manifest

import (
	"sort"
	"time"

	"code-index-plugin/internal/index/model"
)

func Build(records []model.FileRecord) model.Manifest {
	files := make(map[string]model.ManifestFileState, len(records))
	for _, rec := range records {
		files[rec.Path] = model.ManifestFileState{
			Path:     rec.Path,
			Hash:     rec.Hash,
			Size:     rec.Size,
			MTime:    rec.MTime,
			Language: rec.Language,
		}
	}
	return model.Manifest{
		Files:       files,
		GeneratedAt: time.Now().Unix(),
	}
}

func Diff(previous, current model.Manifest) model.ManifestDiff {
	diff := model.ManifestDiff{}
	for path, prev := range previous.Files {
		curr, ok := current.Files[path]
		if !ok {
			diff.Deleted = append(diff.Deleted, path)
			continue
		}
		if curr.Hash != prev.Hash || curr.Size != prev.Size || curr.MTime != prev.MTime {
			diff.Changed = append(diff.Changed, path)
			continue
		}
		diff.Unchanged = append(diff.Unchanged, path)
	}
	for path := range current.Files {
		if _, ok := previous.Files[path]; !ok {
			diff.Added = append(diff.Added, path)
		}
	}
	sort.Strings(diff.Added)
	sort.Strings(diff.Changed)
	sort.Strings(diff.Deleted)
	sort.Strings(diff.Unchanged)
	return diff
}
