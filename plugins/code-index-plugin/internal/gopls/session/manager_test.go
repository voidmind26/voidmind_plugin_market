package session

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

type fakeClient struct {
	closed atomic.Int32
}

func (c *fakeClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}

func (c *fakeClient) Close() error {
	c.closed.Add(1)
	return nil
}

func TestManagerReusesConcurrentWorkspaceSession(t *testing.T) {
	var starts atomic.Int32
	created := &fakeClient{}
	releaseFactory := make(chan struct{})
	manager := NewManager(func(context.Context, string) (Client, error) {
		starts.Add(1)
		<-releaseFactory
		return created, nil
	})

	const callers = 12
	var wg sync.WaitGroup
	results := make(chan Client, callers)
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := manager.Get(context.Background(), "/repo")
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}
			results <- got
		}()
	}
	close(releaseFactory)
	wg.Wait()
	close(results)

	if starts.Load() != 1 {
		t.Fatalf("factory starts = %d, want 1", starts.Load())
	}
	for got := range results {
		if got != created {
			t.Fatalf("Get() returned %p, want %p", got, created)
		}
	}
}

func TestManagerRecreatesInvalidatedSession(t *testing.T) {
	var clients []*fakeClient
	manager := NewManager(func(context.Context, string) (Client, error) {
		created := &fakeClient{}
		clients = append(clients, created)
		return created, nil
	})

	first, err := manager.Get(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("first Get() error = %v", err)
	}
	manager.Invalidate("/repo", first)
	second, err := manager.Get(context.Background(), "/repo")
	if err != nil {
		t.Fatalf("second Get() error = %v", err)
	}
	if first == second || len(clients) != 2 {
		t.Fatalf("invalidated session was reused: first=%p second=%p starts=%d", first, second, len(clients))
	}
	if clients[0].closed.Load() != 1 {
		t.Fatalf("invalidated client close count = %d, want 1", clients[0].closed.Load())
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if clients[1].closed.Load() != 1 {
		t.Fatalf("active client close count = %d, want 1", clients[1].closed.Load())
	}
}
