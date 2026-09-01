package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"gateway-platform-plugin/internal/buildinfo"
	gatewaymcp "gateway-platform-plugin/internal/mcp"
	"gateway-platform-plugin/internal/platformdata"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	pluginRoot := "."
	baseURL := "http://127.0.0.1:18787"
	consoleURL := baseURL + "/app"
	paths, err := platformdata.Prepare(pluginRoot)
	if err != nil {
		log.Fatalf("准备 gateway platform 数据目录失败: %v", err)
	}
	httpExecutable, err := filepath.Abs(filepath.Join(pluginRoot, "bin", "gateway-platform-http"))
	if err != nil {
		log.Fatalf("解析 gateway platform HTTP 程序路径失败: %v", err)
	}

	client := gatewaymcp.NewClient(baseURL, &http.Client{Timeout: 2 * time.Second})
	runtime := gatewaymcp.Runtime{
		BaseURL:         baseURL,
		Client:          client,
		Start:           gatewaymcp.StartHTTPPlatformCommand(httpExecutable, paths.Log),
		Stop:            gatewaymcp.StopHTTPPlatformCommand(18787),
		PollInterval:    100 * time.Millisecond,
		Timeout:         5 * time.Second,
		ExpectedDataDir: paths.Root,
		ExpectedVersion: buildinfo.Version,
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
