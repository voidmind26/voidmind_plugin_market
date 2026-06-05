package query

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"code-index-plugin/internal/index/model"
)

var wordPattern = regexp.MustCompile(`[A-Za-z][A-Za-z0-9_]{2,}`)

type SearchRequest struct {
	Query          string
	Limit          int
	PreferDeepHits bool
	PathPrefix     string
}

type SearchHit struct {
	Kind        string  `json:"kind"`
	Path        string  `json:"path"`
	Title       string  `json:"title"`
	StartLine   int     `json:"start_line,omitempty"`
	EndLine     int     `json:"end_line,omitempty"`
	Summary     string  `json:"summary,omitempty"`
	Score       float64 `json:"score"`
	ScoreReason string  `json:"score_reason,omitempty"`
}

type SearchResponse struct {
	ResultCount int         `json:"result_count"`
	Results     []SearchHit `json:"results"`
}

type scoredHit struct {
	SearchHit
	MatchedTerms int
}

type weightedField struct {
	name   string
	text   string
	weight float64
}

func Search(payload model.ProjectIndex, req SearchRequest) SearchResponse {
	terms := normalizeTerms(req.Query)
	if len(terms) == 0 {
		return SearchResponse{Results: []SearchHit{}}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	var scored []scoredHit
	scored = append(scored, scoreSymbolHits(payload.Symbols, terms, req)...)
	scored = append(scored, scoreChunkHits(payload.Chunks, terms, req)...)
	scored = append(scored, scoreFileHits(payload.Files, terms, req)...)

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].MatchedTerms == scored[j].MatchedTerms {
			if scored[i].Score == scored[j].Score {
				if scored[i].Path == scored[j].Path {
					if scored[i].StartLine == scored[j].StartLine {
						return scored[i].Title < scored[j].Title
					}
					return scored[i].StartLine < scored[j].StartLine
				}
				return scored[i].Path < scored[j].Path
			}
			return scored[i].Score > scored[j].Score
		}
		return scored[i].MatchedTerms > scored[j].MatchedTerms
	})

	results := make([]SearchHit, 0, min(limit, len(scored)))
	for idx, hit := range scored {
		if idx >= limit {
			break
		}
		results = append(results, hit.SearchHit)
	}

	return SearchResponse{
		ResultCount: len(results),
		Results:     results,
	}
}

func scoreSymbolHits(symbols []model.SymbolRecord, terms []string, req SearchRequest) []scoredHit {
	results := make([]scoredHit, 0, len(symbols))
	for _, symbol := range symbols {
		if !matchesPathPrefix(symbol.Path, req.PathPrefix) {
			continue
		}
		title := symbol.SymbolName
		if symbol.Receiver != "" {
			title = symbol.Receiver + "." + symbol.SymbolName
		}
		score, matchedTerms, reason := scoreText(terms,
			weightedField{name: "path", text: symbol.Path, weight: 2},
			weightedField{name: "title", text: title, weight: 8},
			weightedField{name: "summary", text: symbol.Summary, weight: 6},
			weightedField{name: "keywords", text: strings.Join(symbol.Keywords, " "), weight: 5},
		)
		if score == 0 {
			continue
		}
		if req.PreferDeepHits {
			score += 30
			reason = "deep symbol hit; " + reason
		}
		results = append(results, scoredHit{MatchedTerms: matchedTerms, SearchHit: SearchHit{
			Kind:        "symbol",
			Path:        symbol.Path,
			Title:       title,
			StartLine:   symbol.StartLine,
			EndLine:     symbol.EndLine,
			Summary:     symbol.Summary,
			Score:       score,
			ScoreReason: reason,
		}})
	}
	return results
}

func scoreChunkHits(chunks []model.ChunkRecord, terms []string, req SearchRequest) []scoredHit {
	results := make([]scoredHit, 0, len(chunks))
	for _, chunk := range chunks {
		if !matchesPathPrefix(chunk.Path, req.PathPrefix) {
			continue
		}
		score, matchedTerms, reason := scoreText(terms,
			weightedField{name: "path", text: chunk.Path, weight: 2},
			weightedField{name: "title", text: chunk.Title, weight: 7},
			weightedField{name: "summary", text: chunk.Summary, weight: 5},
			weightedField{name: "keywords", text: strings.Join(chunk.Keywords, " "), weight: 4},
		)
		if score == 0 {
			continue
		}
		if req.PreferDeepHits {
			score += 15
			reason = "deep chunk hit; " + reason
		}
		results = append(results, scoredHit{MatchedTerms: matchedTerms, SearchHit: SearchHit{
			Kind:        "chunk",
			Path:        chunk.Path,
			Title:       chunk.Title,
			StartLine:   chunk.StartLine,
			EndLine:     chunk.EndLine,
			Summary:     chunk.Summary,
			Score:       score,
			ScoreReason: reason,
		}})
	}
	return results
}

func scoreFileHits(files []model.FileRecord, terms []string, req SearchRequest) []scoredHit {
	results := make([]scoredHit, 0, len(files))
	for _, file := range files {
		if !matchesPathPrefix(file.Path, req.PathPrefix) {
			continue
		}
		title := filepath.Base(file.Path)
		score, matchedTerms, reason := scoreText(terms,
			weightedField{name: "path", text: file.Path, weight: 2},
			weightedField{name: "title", text: title, weight: 6},
			weightedField{name: "summary", text: file.Summary, weight: 4},
			weightedField{name: "keywords", text: strings.Join(file.Keywords, " "), weight: 4},
		)
		if score == 0 {
			continue
		}
		results = append(results, scoredHit{MatchedTerms: matchedTerms, SearchHit: SearchHit{
			Kind:        "file",
			Path:        file.Path,
			Title:       title,
			Summary:     file.Summary,
			Score:       score,
			ScoreReason: reason,
		}})
	}
	return results
}

func scoreText(terms []string, fields ...weightedField) (float64, int, string) {
	score := 0.0
	matchedTerms := 0
	reasons := make([]string, 0, len(terms))
	for _, term := range terms {
		bestWeight := 0.0
		bestField := ""
		for _, field := range fields {
			if field.text == "" {
				continue
			}
			for _, candidate := range splitTerms(field.text) {
				if candidate != term || field.weight <= bestWeight {
					continue
				}
				bestWeight = field.weight
				bestField = field.name
			}
		}
		if bestWeight == 0 {
			continue
		}
		matchedTerms++
		score += bestWeight * 10
		reasons = append(reasons, term+"@"+bestField)
	}
	if matchedTerms == 0 {
		return 0, 0, ""
	}
	return score + float64(matchedTerms), matchedTerms, "matched terms: " + strings.Join(reasons, ", ")
}

func normalizeTerms(query string) []string {
	return splitTerms(query)
}

func splitTerms(text string) []string {
	matches := wordPattern.FindAllString(text, -1)
	seen := make(map[string]struct{}, len(matches))
	terms := make([]string, 0, len(matches)*2)
	for _, match := range matches {
		for _, part := range splitIdentifierTerms(match) {
			normalized := strings.ToLower(strings.Trim(part, "_"))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			terms = append(terms, normalized)
		}
	}
	return terms
}

func splitIdentifierTerms(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == '-' || r == '/' || r == '.' || r == ':'
	})
	out := make([]string, 0, len(parts)*2)
	for _, part := range parts {
		if part == "" {
			continue
		}
		out = append(out, part)
		for _, camel := range splitCamelCase(part) {
			if camel != part {
				out = append(out, camel)
			}
		}
	}
	return out
}

func splitCamelCase(s string) []string {
	if s == "" {
		return nil
	}
	runes := []rune(s)
	start := 0
	parts := make([]string, 0, len(runes)/2+1)
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		curr := runes[i]
		nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
		boundary := (unicode.IsLower(prev) && unicode.IsUpper(curr)) ||
			(unicode.IsLetter(prev) && unicode.IsDigit(curr)) ||
			(unicode.IsDigit(prev) && unicode.IsLetter(curr)) ||
			(unicode.IsUpper(prev) && unicode.IsUpper(curr) && nextLower)
		if !boundary {
			continue
		}
		parts = append(parts, string(runes[start:i]))
		start = i
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

func matchesPathPrefix(path, prefix string) bool {
	prefix = strings.Trim(filepath.ToSlash(strings.TrimSpace(prefix)), "/")
	if prefix == "" {
		return true
	}
	path = filepath.ToSlash(path)
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
