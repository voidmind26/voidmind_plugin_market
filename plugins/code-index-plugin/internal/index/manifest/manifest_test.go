package manifest

import (
	"reflect"
	"testing"

	"code-index-plugin/internal/index/model"
)

func TestDiffManifestDetectsAddedChangedDeletedFiles(t *testing.T) {
	previous := model.Manifest{
		Files: map[string]model.ManifestFileState{
			"a.go":       {Hash: "old"},
			"b.go":       {Hash: "same"},
			"deleted.ts": {Hash: "gone"},
		},
	}
	current := Build([]model.FileRecord{
		{Path: "a.go", Hash: "new"},
		{Path: "b.go", Hash: "same"},
		{Path: "c.go", Hash: "fresh"},
	})

	diff := Diff(previous, current)

	if !reflect.DeepEqual([]string{"c.go"}, diff.Added) {
		t.Fatalf("expected added [c.go], got %v", diff.Added)
	}
	if !reflect.DeepEqual([]string{"a.go"}, diff.Changed) {
		t.Fatalf("expected changed [a.go], got %v", diff.Changed)
	}
	if !reflect.DeepEqual([]string{"deleted.ts"}, diff.Deleted) {
		t.Fatalf("expected deleted [deleted.ts], got %v", diff.Deleted)
	}
	if !reflect.DeepEqual([]string{"b.go"}, diff.Unchanged) {
		t.Fatalf("expected unchanged [b.go], got %v", diff.Unchanged)
	}
}
