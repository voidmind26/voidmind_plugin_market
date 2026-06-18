package main

import (
	"log"

	"local-db-access/internal/config"
	appserver "local-db-access/internal/server"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := config.InitConf(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	s, err := appserver.New()
	if err != nil {
		log.Fatalf("MCP服务器创建失败: %v", err)
	}

	log.Println("测试数据库 MCP 服务正在启动（stdio 模式）...")
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("MCP服务器启动失败: %v", err)
	}
}
