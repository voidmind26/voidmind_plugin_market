package router

import (
	"net/http"
	"os"

	"gateway-platform-plugin/internal/buildinfo"
	"gateway-platform-plugin/internal/platformdata"
	"gateway-platform-plugin/server/controllers"
	"gateway-platform-plugin/server/helpers"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *helpers.App) *gin.Engine {
	r := gin.Default()
	r.GET("/api/health", func(c *gin.Context) {
		writeErr := platformdata.CheckWritable(app.DataDir)
		if writeErr == nil {
			writeErr = platformdata.CheckDatabaseWritable(c.Request.Context(), app.DB)
		}
		executablePath, executableErr := os.Executable()
		body := gin.H{
			"ok":                writeErr == nil,
			"data_dir":          app.DataDir,
			"database_path":     app.DatabasePath,
			"database_writable": writeErr == nil,
			"pid":               os.Getpid(),
			"executable_path":   executablePath,
			"version":           buildinfo.Version,
		}
		if executableErr != nil {
			body["ok"] = false
			body["error"] = executableErr.Error()
			c.JSON(http.StatusServiceUnavailable, body)
			return
		}
		if writeErr != nil {
			body["error"] = writeErr.Error()
			c.JSON(http.StatusServiceUnavailable, body)
			return
		}
		c.JSON(http.StatusOK, body)
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
