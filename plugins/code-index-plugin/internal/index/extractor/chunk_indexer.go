package extractor

import (
	"path/filepath"
	"strings"

	"code-index-plugin/internal/index/model"
)

func ExtractHeuristicChunks(path, language string, content []byte) []model.ChunkRecord {
	path = filepath.ToSlash(path)
	lines := strings.Split(string(content), "\n")
	chunks := make([]model.ChunkRecord, 0)
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		chunkType, ok := heuristicChunkType(trimmed)
		if !ok {
			continue
		}

		startLine := idx + 1
		endLine := startLine
		switch chunkType {
		case "markdown_heading":
			endLine = findMarkdownHeadingEnd(lines, idx)
		case "heuristic":
			if opensBraceBlock(trimmed) {
				endLine = findBraceBlockEnd(lines, idx)
			}
		}

		chunks = append(chunks, model.ChunkRecord{
			Path:      path,
			Language:  language,
			ChunkType: chunkType,
			Title:     trimmed,
			StartLine: startLine,
			EndLine:   endLine,
			Summary:   trimmed,
			Keywords:  splitTerms(trimmed),
		})
	}
	return chunks
}

func heuristicChunkType(line string) (string, bool) {
	switch {
	case strings.HasPrefix(line, "export "):
		return "heuristic", true
	case strings.HasPrefix(line, "func "):
		return "heuristic", true
	case strings.HasPrefix(line, "type "):
		return "heuristic", true
	case isMarkdownHeading(line):
		return "markdown_heading", true
	default:
		return "", false
	}
}

func isMarkdownHeading(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	trimmed := strings.TrimLeft(line, "#")
	if trimmed == "" {
		return true
	}
	return strings.HasPrefix(trimmed, " ")
}

func opensBraceBlock(line string) bool {
	return strings.HasSuffix(line, "{") && !strings.Contains(line, "}")
}

func findBraceBlockEnd(lines []string, startIdx int) int {
	depth := 0
	for idx := startIdx; idx < len(lines); idx++ {
		line := lines[idx]
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")
		if depth <= 0 {
			return idx + 1
		}
	}
	return len(lines)
}

func findMarkdownHeadingEnd(lines []string, startIdx int) int {
	for idx := startIdx + 1; idx < len(lines); idx++ {
		if isMarkdownHeading(strings.TrimSpace(lines[idx])) {
			return idx
		}
	}
	for idx := len(lines) - 1; idx >= startIdx; idx-- {
		if strings.TrimSpace(lines[idx]) != "" {
			return idx + 1
		}
	}
	return startIdx + 1
}
