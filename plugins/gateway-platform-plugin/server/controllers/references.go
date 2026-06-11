package controllers

import (
	"net/http"

	"gateway-platform-plugin/server/data"
	"gateway-platform-plugin/server/helpers"

	"github.com/gin-gonic/gin"
)

func ListReferences(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		missing, err := data.ListMissingReferences(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		unused, err := data.ListUnusedKeys(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"missing_references": missing,
			"unused_keys":        unused,
		})
	}
}
