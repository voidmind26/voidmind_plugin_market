package index

import (
	"context"

	indexservice "code-index-plugin/internal/index/service"
)

type Service interface {
	Build(context.Context, indexservice.BuildRequest) (indexservice.BuildResult, error)
	Refresh(context.Context, indexservice.RefreshRequest) (indexservice.RefreshResult, error)
	Search(context.Context, indexservice.SearchRequest) (indexservice.SearchResult, error)
	Status(context.Context, indexservice.StatusRequest) (indexservice.StatusResult, error)
}
