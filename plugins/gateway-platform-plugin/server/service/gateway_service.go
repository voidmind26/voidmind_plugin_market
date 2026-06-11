package service

import (
	"net/http"
	"strings"

	"gateway-platform-plugin/server/models"
)

type KeyLike struct {
	ID    int64
	Name  string
	Value string
}

type RewriteLike struct {
	RewriteType string
	TargetName  string
	KeyID       int64
	Template    string
}

func ApplyRewritesLike(req *http.Request, rewrites []RewriteLike, keys map[int64]KeyLike) error {
	for _, rewrite := range rewrites {
		key := keys[rewrite.KeyID]
		value := strings.ReplaceAll(rewrite.Template, "{{value}}", key.Value)
		switch rewrite.RewriteType {
		case "header":
			req.Header.Set(rewrite.TargetName, value)
		case "query":
			q := req.URL.Query()
			q.Set(rewrite.TargetName, value)
			req.URL.RawQuery = q.Encode()
		case "cookie":
			req.AddCookie(&http.Cookie{Name: rewrite.TargetName, Value: value})
		}
	}
	return nil
}

func MatchRouteByPrefix(routes []models.Route, path string) (*models.Route, string) {
	var matched *models.Route
	var rest string
	longest := -1
	for i := range routes {
		prefix := strings.TrimPrefix(routes[i].LocalPath, "/")
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			if len(prefix) > longest {
				matched = &routes[i]
				longest = len(prefix)
				rest = strings.TrimPrefix(path, prefix)
				if rest != "" && !strings.HasPrefix(rest, "/") {
					rest = "/" + rest
				}
			}
		}
	}
	return matched, rest
}

func JoinUpstreamURL(base string, rest string) string {
	if rest == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(rest, "/")
}
