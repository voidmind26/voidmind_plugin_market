package mcp

import "context"

func EnsurePlatformAtStartup(ctx context.Context, runtime Runtime) error {
	if runtime.Start == nil {
		return nil
	}
	return runtime.Start(ctx)
}
