package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client interface {
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	Close() error
}

type Factory func(context.Context, string) (Client, error)

type entry struct {
	ready  chan struct{}
	client Client
	err    error
}

type Manager struct {
	mu      sync.Mutex
	factory Factory
	entries map[string]*entry
	closed  bool
}

func NewManager(factory Factory) *Manager {
	return &Manager{
		factory: factory,
		entries: make(map[string]*entry),
	}
}

func (m *Manager) Get(ctx context.Context, root string) (Client, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, errors.New("gopls 会话管理器已关闭")
	}
	if existing, ok := m.entries[root]; ok {
		m.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-existing.ready:
			return existing.client, existing.err
		}
	}

	created := &entry{ready: make(chan struct{})}
	m.entries[root] = created
	m.mu.Unlock()

	created.client, created.err = m.factory(ctx, root)

	m.mu.Lock()
	if created.err != nil {
		delete(m.entries, root)
	}
	closed := m.closed
	close(created.ready)
	m.mu.Unlock()

	if closed && created.client != nil {
		return nil, errors.New("gopls 会话管理器已关闭")
	}
	return created.client, created.err
}

func (m *Manager) Active(root string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.entries[root]
	if !ok {
		return false
	}
	select {
	case <-current.ready:
		return current.client != nil && current.err == nil
	default:
		return false
	}
}

func (m *Manager) Invalidate(root string, failed Client) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	current, ok := m.entries[root]
	if !ok || current.client != failed {
		m.mu.Unlock()
		return
	}
	delete(m.entries, root)
	m.mu.Unlock()

	_ = failed.Close()
}

func (m *Manager) Close() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	entries := make([]*entry, 0, len(m.entries))
	for _, current := range m.entries {
		entries = append(entries, current)
	}
	m.mu.Unlock()

	var closeErrors []error
	for _, current := range entries {
		<-current.ready
		if current.client != nil {
			if err := current.client.Close(); err != nil {
				closeErrors = append(closeErrors, err)
			}
		}
	}
	return errors.Join(closeErrors...)
}

func NewGoplsClient(ctx context.Context, root string) (Client, error) {
	goplsPath, err := exec.LookPath("gopls")
	if err != nil {
		return nil, errors.New("gopls 未安装或不在 PATH 中，请执行 go install golang.org/x/tools/gopls@latest")
	}
	if err := exec.CommandContext(ctx, goplsPath, "help", "mcp").Run(); err != nil {
		return nil, fmt.Errorf("当前 gopls 不支持 mcp 子命令，请升级到 v0.22.0 或更高版本: %w", err)
	}

	const launch = `unset GOWORK
cd -- "$1" || exit 1
exec "$2" mcp`
	c, err := client.NewStdioMCPClient("bash", nil, "-lc", launch, "gopls-router", root, goplsPath)
	if err != nil {
		return nil, fmt.Errorf("在 %s 启动 gopls mcp 失败: %w", root, err)
	}
	if stderr, ok := client.GetStderr(c); ok {
		go func() {
			_, _ = io.Copy(io.Discard, stderr)
		}()
	}

	request := mcp.InitializeRequest{}
	request.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = mcp.Implementation{
		Name:    "code-index-gopls-router",
		Version: "0.2.0",
	}
	if _, err := c.Initialize(ctx, request); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("初始化 %s 的 gopls mcp 失败: %w", root, err)
	}
	return c, nil
}
