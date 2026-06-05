package query

import (
	"testing"

	"code-index-plugin/internal/index/model"
)

func TestSearchPrefersDeepSymbolHitsOverFileOnlyHits(t *testing.T) {
	payload := model.ProjectIndex{
		Files: []model.FileRecord{
			{Path: "docs/payment.md", Summary: "payment callback overview", Keywords: []string{"payment", "callback"}},
			{Path: "service/payment.go", Summary: "payment callback service", Keywords: []string{"payment", "callback"}},
		},
		Symbols: []model.SymbolRecord{
			{Path: "service/payment.go", SymbolName: "HandleCallback", Summary: "handle payment callback", StartLine: 10, EndLine: 20, Keywords: []string{"payment", "callback"}},
		},
	}

	got := Search(payload, SearchRequest{Query: "payment callback", Limit: 5, PreferDeepHits: true})
	if len(got.Results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if got.Results[0].Kind != "symbol" {
		t.Fatalf("expected first result kind symbol, got %q", got.Results[0].Kind)
	}
	if got.Results[0].Title != "HandleCallback" {
		t.Fatalf("expected first result title HandleCallback, got %q", got.Results[0].Title)
	}
}

func TestSearchIncludesFileAndChunkMatchesAndHonorsLimit(t *testing.T) {
	payload := model.ProjectIndex{
		Files: []model.FileRecord{
			{Path: "service/user.go", Summary: "user profile handlers", Keywords: []string{"user", "profile"}},
		},
		Chunks: []model.ChunkRecord{
			{Path: "web/profile.ts", Title: "export function renderUserProfile()", Summary: "render user profile", StartLine: 3, EndLine: 8, Keywords: []string{"user", "profile"}},
		},
	}

	got := Search(payload, SearchRequest{Query: "user profile", Limit: 1})
	if got.ResultCount != 1 {
		t.Fatalf("expected result count 1, got %d", got.ResultCount)
	}
	if len(got.Results) != 1 {
		t.Fatalf("expected one limited result, got %d", len(got.Results))
	}
	if got.Results[0].Kind != "chunk" && got.Results[0].Kind != "file" {
		t.Fatalf("expected file or chunk result kind, got %q", got.Results[0].Kind)
	}
}

func TestSearchPrefersCoverageOverFolderNameNoise(t *testing.T) {
	payload := model.ProjectIndex{
		Files: []model.FileRecord{
			{Path: "controllers/command/test.go", Summary: "command tests", Keywords: []string{"controllers"}},
			{Path: "helpers/login.go", Summary: "login user helpers", Keywords: []string{"login", "user"}},
		},
		Symbols: []model.SymbolRecord{
			{Path: "helpers/login.go", SymbolName: "GetLoginUserInfo", Summary: "get login user info", StartLine: 10, EndLine: 20, Keywords: []string{"login", "user", "info"}},
		},
	}

	got := Search(payload, SearchRequest{Query: "controllers user", Limit: 3, PreferDeepHits: true})
	if len(got.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if got.Results[0].Path != "helpers/login.go" {
		t.Fatalf("expected user-related result first, got %+v", got.Results[0])
	}
}

func TestSearchDoesNotMatchShortTermAsSubstringInsideAnotherWord(t *testing.T) {
	payload := model.ProjectIndex{
		Symbols: []model.SymbolRecord{
			{Path: "service/srvhpa/srv.go", SymbolName: "CalTargetReplicasByUnUseReplicas", Summary: "calculate target replicas", StartLine: 1, EndLine: 2, Keywords: []string{"cal", "target", "replicas", "un", "use", "replicas"}},
			{Path: "helpers/login.go", SymbolName: "GetLoginUserInfo", Summary: "get login user info", StartLine: 10, EndLine: 20, Keywords: []string{"login", "user", "info"}},
		},
	}

	got := Search(payload, SearchRequest{Query: "user", Limit: 3, PreferDeepHits: true})
	if len(got.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if got.Results[0].Path != "helpers/login.go" {
		t.Fatalf("expected exact user term match first, got %+v", got.Results[0])
	}
}
