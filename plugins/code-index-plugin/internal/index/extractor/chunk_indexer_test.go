package extractor

import (
	"testing"

	"code-index-plugin/internal/index/model"
)

func TestExtractHeuristicChunksReturnsExportBlocks(t *testing.T) {
	content := []byte(`# Users
Overview line
export function loadUsers() {}
const localOnly = true

## Types
type UserRow = {
	id: string
}

func helper() {}
export const columns = []
`)

	chunks := ExtractHeuristicChunks("src/users.ts", "typescript", content)
	if len(chunks) != 6 {
		t.Fatalf("expected 6 chunks, got %d (%v)", len(chunks), chunks)
	}

	assertChunk(t, chunks[0], "markdown_heading", "# Users", 1, 5)
	assertContains(t, chunks[0].Keywords, "users")

	assertChunk(t, chunks[1], "heuristic", "export function loadUsers() {}", 3, 3)
	assertContains(t, chunks[1].Keywords, "export")
	assertContains(t, chunks[1].Keywords, "loadusers")

	assertChunk(t, chunks[2], "markdown_heading", "## Types", 6, 12)
	assertContains(t, chunks[2].Keywords, "types")

	assertChunk(t, chunks[3], "heuristic", "type UserRow = {", 7, 9)
	assertContains(t, chunks[3].Keywords, "type")

	assertChunk(t, chunks[4], "heuristic", "func helper() {}", 11, 11)
	assertContains(t, chunks[4].Keywords, "helper")

	assertChunk(t, chunks[5], "heuristic", "export const columns = []", 12, 12)
}

func assertChunk(t *testing.T, got model.ChunkRecord, wantType, wantTitle string, wantStartLine, wantEndLine int) {
	t.Helper()
	if got.ChunkType != wantType {
		t.Fatalf("expected chunk type %q, got %q", wantType, got.ChunkType)
	}
	if got.Title != wantTitle {
		t.Fatalf("expected title %q, got %q", wantTitle, got.Title)
	}
	if got.StartLine != wantStartLine || got.EndLine != wantEndLine {
		t.Fatalf("expected lines %d-%d, got %d-%d", wantStartLine, wantEndLine, got.StartLine, got.EndLine)
	}
}
