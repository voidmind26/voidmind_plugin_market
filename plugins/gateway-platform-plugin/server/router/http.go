package router

import (
	"net/http"

	"gateway-platform-plugin/server/controllers"
	"gateway-platform-plugin/server/helpers"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *helpers.App) *gin.Engine {
	r := gin.Default()
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/routes", controllers.ListRoutes(app))
	r.POST("/api/routes", controllers.CreateRoute(app))
	r.GET("/api/routes/:id", controllers.GetRoute(app))
	r.PUT("/api/routes/:id", controllers.UpdateRoute(app))
	r.DELETE("/api/routes/:id", controllers.DeleteRoute(app))
	r.GET("/api/keys", controllers.ListKeys(app))
	r.POST("/api/keys", controllers.CreateKey(app))
	r.GET("/api/keys/:id", controllers.GetKey(app))
	r.PUT("/api/keys/:id", controllers.UpdateKey(app))
	r.DELETE("/api/keys/:id", controllers.DeleteKey(app))
	r.GET("/api/routes/:id/rewrites", controllers.ListRewrites(app))
	r.POST("/api/routes/:id/rewrites", controllers.CreateRewrite(app))
	r.PUT("/api/routes/:id/rewrites/:rewriteId", controllers.UpdateRewrite(app))
	r.DELETE("/api/routes/:id/rewrites/:rewriteId", controllers.DeleteRewrite(app))
	r.GET("/api/references", controllers.ListReferences(app))
	registerGatewayRoutes(r, app)
	registerFrontendRoutes(r)
	return r
}

func RunHTTP(app *helpers.App) error {
	return NewRouter(app).Run("127.0.0.1:18787")
}
