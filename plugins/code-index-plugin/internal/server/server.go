package server

import (
	goplsrouter "code-index-plugin/internal/gopls/router"
	"code-index-plugin/internal/gopls/session"
	"code-index-plugin/internal/gopls/workspace"
	"code-index-plugin/internal/index/scanner"
	indexservice "code-index-plugin/internal/index/service"
	"code-index-plugin/internal/index/storage"
	goplstool "code-index-plugin/internal/tools/gopls"
	indextool "code-index-plugin/internal/tools/index"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

type Server struct {
	MCP      *mcpserver.MCPServer
	sessions *session.Manager
}

func New() (*Server, error) {
	mcpServer := mcpserver.NewMCPServer(
		"code-index-plugin",
		"0.2.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithRecovery(),
	)

	svc := indexservice.New(storage.New(), scanner.New(indexservice.DefaultOptions()))
	indexHandler := indextool.NewHandler(svc)
	indexHandler.RegisterTools(mcpServer)

	sessions := session.NewManager(session.NewGoplsClient)
	router := goplsrouter.New(workspace.NewResolver(), sessions)
	goplsHandler := goplstool.NewHandler(router)
	goplsHandler.RegisterTools(mcpServer)

	return &Server{MCP: mcpServer, sessions: sessions}, nil
}

func (s *Server) Close() error {
	return s.sessions.Close()
}
