package main

import (
	"context"
	"log"
	"net/http"
	"time"

	gatewaymcp "gateway-platform-plugin/internal/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	pluginRoot := "."
	baseURL := "http://127.0.0.1:18787"
	consoleURL := baseURL + "/app"

	client := gatewaymcp.NewClient(baseURL, &http.Client{Timeout: 2 * time.Second})
	runtime := gatewaymcp.Runtime{
		BaseURL:      baseURL,
		Client:       client,
		Start:        gatewaymcp.StartHTTPPlatformCommand(pluginRoot),
		PollInterval: 100 * time.Millisecond,
		Timeout:      5 * time.Second,
	}

	if err := gatewaymcp.EnsurePlatformAtStartup(context.Background(), runtime); err != nil {
		log.Printf("gateway platform HTTP 预热失败: %v", err)
	}

	server := gatewaymcp.NewToolServerWithRuntime(client, consoleURL, runtime).Build()

	log.Println("gateway platform MCP 服务正在启动（stdio 模式）...")
	if err := mcpserver.ServeStdio(server); err != nil {
		log.Fatalf("MCP服务器启动失败: %v", err)
	}
}
