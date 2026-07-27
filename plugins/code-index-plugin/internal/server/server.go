package server

import (
	"code-index-plugin/internal/index/scanner"
	indexservice "code-index-plugin/internal/index/service"
	"code-index-plugin/internal/index/storage"
	indextool "code-index-plugin/internal/tools/index"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func New() (*mcpserver.MCPServer, error) {
	s := mcpserver.NewMCPServer(
		"code-index-plugin",
		"0.2.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithRecovery(),
	)

	svc := indexservice.New(storage.New(), scanner.New(indexservice.DefaultOptions()))
	handler := indextool.NewHandler(svc)
	handler.RegisterTools(s)

	return s, nil
}
