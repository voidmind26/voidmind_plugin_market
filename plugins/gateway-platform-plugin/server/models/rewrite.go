package models

type RouteRewrite struct {
	ID          int64  `json:"id"`
	RouteID     int64  `json:"route_id"`
	RewriteType string `json:"rewrite_type"`
	TargetName  string `json:"target_name"`
	KeyID       int64  `json:"key_id"`
	Template    string `json:"template"`
	Ordering    int    `json:"ordering"`
}
