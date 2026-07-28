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
	if !manager.Active("/repo") {
		t.Fatal("Active() = false after successful initialization")
	}
}

func TestManagerCloseClosesEverySessionOnce(t *testing.T) {
	clients := map[string]*fakeClient{}
	manager := NewManager(func(_ context.Context, root string) (Client, error) {
		client := &fakeClient{}
		clients[root] = client
		return client, nil
	})

	for _, root := range []string{"/repo-a", "/repo-b"} {
		if _, err := manager.Get(context.Background(), root); err != nil {
			t.Fatalf("Get(%s) error = %v", root, err)
		}
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	for root, client := range clients {
		if client.closed.Load() != 1 {
			t.Errorf("client %s close count = %d, want 1", root, client.closed.Load())
		}
	}
}
