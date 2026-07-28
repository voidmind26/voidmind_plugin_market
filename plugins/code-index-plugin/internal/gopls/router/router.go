package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"code-index-plugin/internal/gopls/session"
	"code-index-plugin/internal/gopls/workspace"

	"github.com/mark3labs/mcp-go/mcp"
)

const defaultFanoutLimit = 4

type Resolver interface {
	Discover(string) ([]workspace.Workspace, error)
	Resolve(string) (workspace.Workspace, error)
	ResolveFile(string) (workspace.Workspace, error)
}

type Sessions interface {
	Get(context.Context, string) (session.Client, error)
	Active(string) bool
	Invalidate(string, session.Client)
}

type Router struct {
	resolver    Resolver
	sessions    Sessions
	fanoutLimit int
}

type WorkspaceStatus struct {
	workspace.Workspace
	SessionActive bool `json:"session_active"`
}

type WorkspaceList struct {
	ProjectRoot string            `json:"project_root"`
	Count       int               `json:"count"`
	Workspaces  []WorkspaceStatus `json:"workspaces"`
}

func New(resolver Resolver, sessions Sessions) *Router {
	return &Router{
		resolver:    resolver,
		sessions:    sessions,
		fanoutLimit: defaultFanoutLimit,
	}
}

func (r *Router) ListWorkspaces(projectRoot string) (WorkspaceList, error) {
	root, err := workspace.Normalize(projectRoot)
	if err != nil {
		return WorkspaceList{}, err
	}
	workspaces, err := r.resolver.Discover(root)
	if err != nil {
		return WorkspaceList{}, err
	}
	statuses := make([]WorkspaceStatus, 0, len(workspaces))
	for _, ws := range workspaces {
		statuses = append(statuses, WorkspaceStatus{
			Workspace:     ws,
			SessionActive: r.sessions.Active(ws.Root),
		})
	}
	return WorkspaceList{ProjectRoot: root, Count: len(statuses), Workspaces: statuses}, nil
}

func (r *Router) CallFanout(
	ctx context.Context,
	toolName string,
	projectRoot string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	workspaces, err := r.resolver.Discover(projectRoot)
	if err != nil {
		return nil, err
	}
	if len(workspaces) == 0 {
		return nil, fmt.Errorf("%w: %s", workspace.ErrNoWorkspace, projectRoot)
	}
	if len(workspaces) == 1 {
		return r.callWorkspace(ctx, workspaces[0], toolName, arguments)
	}
	return r.callMany(ctx, workspaces, toolName, arguments), nil
}

func (r *Router) CallSingle(
	ctx context.Context,
	toolName string,
	projectRoot string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	workspaces, err := r.resolver.Discover(projectRoot)
	if err != nil {
		return nil, err
	}
	if len(workspaces) == 0 {
		return nil, fmt.Errorf("%w: %s", workspace.ErrNoWorkspace, projectRoot)
	}
	if len(workspaces) > 1 {
		return nil, fmt.Errorf("%s 需要明确的单个 Go workspace，%s 下发现了 %d 个；请将 project_root 指向具体仓库", toolName, projectRoot, len(workspaces))
	}
	return r.callWorkspace(ctx, workspaces[0], toolName, arguments)
}

func (r *Router) CallFile(
	ctx context.Context,
	toolName string,
	file string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	ws, err := r.resolver.ResolveFile(file)
	if err != nil {
		return nil, err
	}
	return r.callWorkspace(ctx, ws, toolName, arguments)
}

func (r *Router) Diagnostics(
	ctx context.Context,
	projectRoot string,
	files []string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	if len(files) == 0 {
		return r.CallFanout(ctx, "go_diagnostics", projectRoot, arguments)
	}

	allowedWorkspaces := make(map[string]bool)
	if projectRoot != "" {
		discovered, err := r.resolver.Discover(projectRoot)
		if err != nil {
			return nil, err
		}
		for _, ws := range discovered {
			allowedWorkspaces[ws.Root] = true
		}
	}

	type group struct {
		workspace workspace.Workspace
		files     []string
	}
	groups := make(map[string]*group)
	for _, file := range files {
		if !filepath.IsAbs(file) {
			return nil, fmt.Errorf("诊断文件路径必须是绝对路径: %s", file)
		}
		absolute := filepath.Clean(file)
		ws, err := r.resolver.ResolveFile(absolute)
		if err != nil {
			return nil, err
		}
		if len(allowedWorkspaces) > 0 && !allowedWorkspaces[ws.Root] {
			return nil, fmt.Errorf("诊断文件解析到 project_root 之外的 workspace: %s", file)
		}
		current := groups[ws.Root]
		if current == nil {
			current = &group{workspace: ws}
			groups[ws.Root] = current
		}
		current.files = append(current.files, absolute)
	}

	workspaces := make([]workspace.Workspace, 0, len(groups))
	perWorkspace := make(map[string]map[string]any, len(groups))
	for root, current := range groups {
		workspaces = append(workspaces, current.workspace)
		args := cloneArguments(arguments)
		args["files"] = current.files
		perWorkspace[root] = args
	}
	if len(workspaces) == 1 {
		return r.callWorkspace(ctx, workspaces[0], "go_diagnostics", perWorkspace[workspaces[0].Root])
	}
	return r.callManyWithArguments(ctx, workspaces, "go_diagnostics", perWorkspace), nil
}

func (r *Router) Vulncheck(
	ctx context.Context,
	projectRoot string,
	dir string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	workspaces, err := r.resolver.Discover(projectRoot)
	if err != nil {
		return nil, err
	}
	if len(workspaces) != 1 {
		return nil, fmt.Errorf("go_vulncheck 需要明确的单个 Go workspace，%s 下发现了 %d 个", projectRoot, len(workspaces))
	}
	if dir != "" {
		normalized, err := workspace.Normalize(dir)
		if err != nil {
			return nil, err
		}
		if !workspace.Contains(workspaces[0].Root, normalized) {
			return nil, fmt.Errorf("漏洞检查目录超出 workspace: %s", dir)
		}
		arguments = cloneArguments(arguments)
		arguments["dir"] = normalized
	}
	return r.callWorkspace(ctx, workspaces[0], "go_vulncheck", arguments)
}

func (r *Router) callWorkspace(
	ctx context.Context,
	ws workspace.Workspace,
	toolName string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	client, err := r.sessions.Get(ctx, ws.Root)
	if err != nil {
		return nil, err
	}
	request := mcp.CallToolRequest{}
	request.Params.Name = toolName
	request.Params.Arguments = cloneArguments(arguments)
	result, err := client.CallTool(ctx, request)
	if err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			r.sessions.Invalidate(ws.Root, client)
		}
		return nil, fmt.Errorf("%s 调用 %s 失败: %w", ws.Root, toolName, err)
	}
	return result, nil
}

type outcome struct {
	root   string
	result *mcp.CallToolResult
	err    error
}

func (r *Router) callMany(
	ctx context.Context,
	workspaces []workspace.Workspace,
	toolName string,
	arguments map[string]any,
) *mcp.CallToolResult {
	perWorkspace := make(map[string]map[string]any, len(workspaces))
	for _, ws := range workspaces {
		perWorkspace[ws.Root] = arguments
	}
	return r.callManyWithArguments(ctx, workspaces, toolName, perWorkspace)
}

func (r *Router) callManyWithArguments(
	ctx context.Context,
	workspaces []workspace.Workspace,
	toolName string,
	perWorkspace map[string]map[string]any,
) *mcp.CallToolResult {
	outcomes := make(chan outcome, len(workspaces))
	limit := make(chan struct{}, r.fanoutLimit)
	var wg sync.WaitGroup
	for _, ws := range workspaces {
		ws := ws
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				outcomes <- outcome{root: ws.Root, err: ctx.Err()}
				return
			case limit <- struct{}{}:
			}
			defer func() { <-limit }()
			result, err := r.callWorkspace(ctx, ws, toolName, perWorkspace[ws.Root])
			outcomes <- outcome{root: ws.Root, result: result, err: err}
		}()
	}
	wg.Wait()
	close(outcomes)

	collected := make([]outcome, 0, len(workspaces))
	for current := range outcomes {
		collected = append(collected, current)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].root < collected[j].root })
	if toolName == "go_search" {
		return aggregateSearch(collected)
	}
	return aggregateText(toolName, collected)
}

func aggregateSearch(outcomes []outcome) *mcp.CallToolResult {
	matchedRoots := make(map[string][]string)
	var matchOrder []string
	var failures []string
	for _, current := range outcomes {
		text, err := outcomeText(current)
		if err != nil {
			failures = append(failures, fmt.Sprintf("- %s: %v", current.root, err))
			continue
		}
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "Top symbol matches:" || line == "No symbols found." {
				continue
			}
			if _, seen := matchedRoots[line]; !seen {
				matchOrder = append(matchOrder, line)
			}
			matchedRoots[line] = append(matchedRoots[line], current.root)
		}
	}

	var output strings.Builder
	if len(matchOrder) == 0 {
		output.WriteString("No symbols found.")
	} else {
		output.WriteString("Top symbol matches across Go workspaces:\n")
		for index, line := range matchOrder {
			if index > 0 {
				output.WriteByte('\n')
			}
			fmt.Fprintf(&output, "\t%s [workspace: %s]", line, strings.Join(matchedRoots[line], ", "))
		}
	}
	appendFailures(&output, failures)
	return aggregatedResult(output.String(), len(outcomes)-len(failures) == 0)
}

func aggregateText(toolName string, outcomes []outcome) *mcp.CallToolResult {
	var output strings.Builder
	var failures []string
	successes := 0
	for _, current := range outcomes {
		text, err := outcomeText(current)
		if err != nil {
			failures = append(failures, fmt.Sprintf("- %s: %v", current.root, err))
			continue
		}
		if successes > 0 {
			output.WriteString("\n\n")
		}
		fmt.Fprintf(&output, "Go workspace `%s`:\n%s", current.root, strings.TrimSpace(text))
		successes++
	}
	if successes == 0 {
		fmt.Fprintf(&output, "%s 未能在任何 Go workspace 中完成。", toolName)
	}
	appendFailures(&output, failures)
	return aggregatedResult(output.String(), successes == 0)
}

func outcomeText(current outcome) (string, error) {
	if current.err != nil {
		return "", current.err
	}
	if current.result == nil {
		return "", errors.New("上游未返回结果")
	}
	var texts []string
	for _, content := range current.result.Content {
		if text, ok := mcp.AsTextContent(content); ok {
			texts = append(texts, text.Text)
		}
	}
	combined := strings.Join(texts, "\n")
	if current.result.IsError {
		if combined == "" {
			combined = "上游返回工具错误"
		}
		return "", errors.New(combined)
	}
	return combined, nil
}

func appendFailures(output *strings.Builder, failures []string) {
	if len(failures) == 0 {
		return
	}
	if output.Len() > 0 {
		output.WriteString("\n\n")
	}
	output.WriteString("未完成的 Go workspace:\n")
	output.WriteString(strings.Join(failures, "\n"))
}

func aggregatedResult(text string, isError bool) *mcp.CallToolResult {
	result := mcp.NewToolResultText(text)
	result.IsError = isError
	return result
}

func cloneArguments(arguments map[string]any) map[string]any {
	cloned := make(map[string]any, len(arguments))
	for key, value := range arguments {
		if key != "project_root" {
			cloned[key] = value
		}
	}
	return cloned
}

func JSONResult(payload any) (*mcp.CallToolResult, error) {
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("格式化 workspace 列表失败: %w", err)
	}
	return mcp.NewToolResultText(string(encoded)), nil
}
