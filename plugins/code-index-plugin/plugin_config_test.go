package codeindexplugin_test

import (
	"encoding/json"
	"os"
	"strings"
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

func TestMCPManifestRegistersWorkspaceAwareGopls(t *testing.T) {
	payload, err := os.ReadFile(".mcp.json")
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}

	var manifest mcpManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		t.Fatalf("decode .mcp.json: %v", err)
	}

	if manifest.Servers["code-index"].CWD != "." {
		t.Fatal("code-index server must run from the plugin root")
	}

	gopls, ok := manifest.Servers["gopls"]
	if !ok {
		t.Fatal("gopls MCP server is not registered")
	}
	if gopls.CWD != "" {
		t.Fatalf("gopls must inherit the consumer workspace, got cwd %q", gopls.CWD)
	}
	if gopls.Command != "bash" || !strings.Contains(strings.Join(gopls.Args, " "), "exec gopls mcp") {
		t.Fatalf("gopls server must launch gopls mcp, got %+v", gopls)
	}
}
