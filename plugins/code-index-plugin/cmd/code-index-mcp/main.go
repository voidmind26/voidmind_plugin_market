package main

import (
	"errors"
	"fmt"
	"log"

	appserver "code-index-plugin/internal/server"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	server, err := appserver.New()
	if err != nil {
		return fmt.Errorf("MCP服务器创建失败: %w", err)
	}
	defer func() {
		if closeErr := server.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("关闭 gopls 会话失败: %w", closeErr))
		}
	}()

	log.Println("代码索引 MCP 服务正在启动（stdio 模式）...")
	if err := mcpserver.ServeStdio(server.MCP); err != nil {
		return fmt.Errorf("MCP服务器运行失败: %w", err)
	}
	return nil
}
