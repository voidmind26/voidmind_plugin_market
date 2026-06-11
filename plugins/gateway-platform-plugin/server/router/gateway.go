package router

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"gateway-platform-plugin/server/data"
	"gateway-platform-plugin/server/helpers"
	"gateway-platform-plugin/server/service"

	"github.com/gin-gonic/gin"
)

func registerGatewayRoutes(r *gin.Engine, app *helpers.App) {
	r.Any("/gateway/*path", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("path"), "/")
		routes, err := data.ListRoutes(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		route, restPath := service.MatchRouteByPrefix(routes, path)
		if route == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		if !route.Enabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "route disabled"})
			return
		}
		rewrites, err := data.ListRewrites(app.DB, route.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		keys, err := data.ListKeys(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		upstreamURL := service.JoinUpstreamURL(route.UpstreamURL, restPath)
		req, err := http.NewRequest(c.Request.Method, upstreamURL, bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		req.Header = c.Request.Header.Clone()
		serviceKeys := map[int64]service.KeyLike{}
		for _, key := range keys {
			serviceKeys[key.ID] = service.KeyLike{ID: key.ID, Name: key.Name, Value: key.Value}
		}
		serviceRewrites := make([]service.RewriteLike, 0, len(rewrites))
		for _, rewrite := range rewrites {
			serviceRewrites = append(serviceRewrites, service.RewriteLike{RewriteType: rewrite.RewriteType, TargetName: rewrite.TargetName, KeyID: rewrite.KeyID, Template: rewrite.Template})
		}
		if err := service.ApplyRewritesLike(req, serviceRewrites, serviceKeys); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}
		c.Status(resp.StatusCode)
		_, _ = io.Copy(c.Writer, resp.Body)
	})
}
