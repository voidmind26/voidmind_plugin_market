package extractor

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"code-index-plugin/internal/index/model"
)

func ExtractGoSymbols(path string, content []byte) ([]model.SymbolRecord, error) {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	path = filepath.ToSlash(path)
	var symbols []model.SymbolRecord
	for _, decl := range parsed.Decls {
		switch typed := decl.(type) {
		case *ast.FuncDecl:
			symbols = append(symbols, buildFuncSymbol(fileSet, path, typed))
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			symbols = append(symbols, buildTypeSymbols(fileSet, path, typed)...)
		}
	}
	return symbols, nil
}

func buildFuncSymbol(fileSet *token.FileSet, path string, decl *ast.FuncDecl) model.SymbolRecord {
	startLine := fileSet.Position(decl.Pos()).Line
	endLine := fileSet.Position(decl.End()).Line
	symbolType := "func"
	receiverType := ""
	if decl.Recv != nil {
		symbolType = "method"
		receiverType = receiverName(decl.Recv)
	}
	summary := docSummary(decl.Doc)
	if summary == "" {
		if receiverType != "" {
			summary = fmt.Sprintf("method %s on %s", decl.Name.Name, receiverType)
		} else {
			summary = fmt.Sprintf("func %s", decl.Name.Name)
		}
	}

	keywords := splitTerms(strings.Join([]string{decl.Name.Name, symbolType, receiverType, summary}, " "))
	return model.SymbolRecord{
		Path:       path,
		Language:   "go",
		SymbolName: decl.Name.Name,
		SymbolType: symbolType,
		Receiver:   receiverType,
		StartLine:  startLine,
		EndLine:    endLine,
		Summary:    summary,
		Keywords:   keywords,
	}
}

func buildTypeSymbols(fileSet *token.FileSet, path string, decl *ast.GenDecl) []model.SymbolRecord {
	var out []model.SymbolRecord
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		startLine := fileSet.Position(typeSpec.Pos()).Line
		endLine := fileSet.Position(typeSpec.End()).Line
		kind := goTypeKind(typeSpec)
		summary := docSummary(typeSpec.Doc)
		if summary == "" {
			summary = docSummary(decl.Doc)
		}
		if summary == "" {
			summary = fmt.Sprintf("%s %s", kind, typeSpec.Name.Name)
		}
		keywords := splitTerms(strings.Join([]string{typeSpec.Name.Name, "type", kind, summary}, " "))
		out = append(out, model.SymbolRecord{
			Path:       path,
			Language:   "go",
			SymbolName: typeSpec.Name.Name,
			SymbolType: "type",
			StartLine:  startLine,
			EndLine:    endLine,
			Summary:    summary,
			Keywords:   keywords,
		})
	}
	return out
}

func goTypeKind(spec *ast.TypeSpec) string {
	if spec.Assign.IsValid() {
		return "alias"
	}
	switch spec.Type.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	default:
		return "type"
	}
}

func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return exprName(recv.List[0].Type)
}

func exprName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return exprName(typed.X)
	case *ast.SelectorExpr:
		left := exprName(typed.X)
		if left == "" {
			return typed.Sel.Name
		}
		return left + "." + typed.Sel.Name
	case *ast.IndexExpr:
		return exprName(typed.X)
	case *ast.IndexListExpr:
		return exprName(typed.X)
	default:
		return ""
	}
}

func docSummary(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	for _, line := range strings.Split(group.Text(), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
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
	sort.Strings(terms)
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
