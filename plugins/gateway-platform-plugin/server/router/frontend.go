package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed frontend_dist/index.html frontend_dist/assets/*
var frontendFS embed.FS

func registerFrontendRoutes(r *gin.Engine) {
	dist, err := fs.Sub(frontendFS, "frontend_dist")
	if err != nil {
		panic(err)
	}
	indexHTML, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		panic(err)
	}
	serveIndex := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}
	r.GET("/app", serveIndex)
	r.GET("/app/*filepath", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("filepath"), "/")
		if path == "" {
			serveIndex(c)
			return
		}
		if strings.HasPrefix(path, "assets/") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".html") {
			c.FileFromFS(path, http.FS(dist))
			return
		}
		serveIndex(c)
	})
}
