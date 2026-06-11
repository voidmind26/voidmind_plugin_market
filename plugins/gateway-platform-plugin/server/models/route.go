package models

type Route struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	LocalPath   string `json:"local_path"`
	UpstreamURL string `json:"upstream_url"`
	TimeoutMS   int    `json:"timeout_ms"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
