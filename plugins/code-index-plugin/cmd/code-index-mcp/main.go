package main

import (
	"log"

	appserver "code-index-plugin/internal/server"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	s, err := appserver.New()
	if err != nil {
		log.Fatalf("MCP服务器创建失败: %v", err)
	}

	log.Println("代码索引 MCP 服务正在启动（stdio 模式）...")
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("MCP服务器启动失败: %v", err)
	}
}
