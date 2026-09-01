package mcp

import "context"

func EnsurePlatformAtStartup(ctx context.Context, runtime Runtime) error {
	if runtime.Client == nil {
		return nil
	}
	return runtime.EnsurePlatform(ctx)
}
