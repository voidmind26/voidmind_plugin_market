package codeindexplugin_test

import (
	"encoding/json"
	"os"
	"testing"
)

type mcpManifest struct {
	Servers map[string]mcpServer `json:"mcpServers"`
}

type mcpServer struct {
	CWD     string   `json:"cwd"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func TestMCPManifestRegistersOnlyRoutedCodeIndexServer(t *testing.T) {
	payload, err := os.ReadFile(".mcp.json")
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}

	var manifest mcpManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatalf("decode .mcp.json: %v", err)
	}

	if len(manifest.Servers) != 1 {
		t.Fatalf("MCP server count = %d, want 1 routed server", len(manifest.Servers))
	}
	codeIndex, ok := manifest.Servers["code-index"]
	if !ok {
		t.Fatal("code-index MCP server is not registered")
	}
	if codeIndex.CWD != "." {
		t.Fatal("code-index server must run from the plugin root")
	}
	if _, exists := manifest.Servers["gopls"]; exists {
		t.Fatal("standalone gopls server must be removed in favor of the multi-workspace router")
	}
}
