package controllers

import (
	"net/http"
	"strconv"

	"gateway-platform-plugin/server/data"
	"gateway-platform-plugin/server/helpers"
	"gateway-platform-plugin/server/models"

	"github.com/gin-gonic/gin"
)

type rewritePayload struct {
	RewriteType string `json:"rewrite_type"`
	TargetName  string `json:"target_name"`
	KeyID       int64  `json:"key_id"`
	Template    string `json:"template"`
	Ordering    int    `json:"ordering"`
}

func ListRewrites(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		routeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		items, err := data.ListRewrites(app.DB, routeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func CreateRewrite(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		routeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var payload rewritePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.CreateRewrite(app.DB, models.RouteRewrite{RouteID: routeID, RewriteType: payload.RewriteType, TargetName: payload.TargetName, KeyID: payload.KeyID, Template: payload.Template, Ordering: payload.Ordering})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	}
}

func UpdateRewrite(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		rewriteID, err := strconv.ParseInt(c.Param("rewriteId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var payload rewritePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.UpdateRewrite(app.DB, rewriteID, models.RouteRewrite{RewriteType: payload.RewriteType, TargetName: payload.TargetName, KeyID: payload.KeyID, Template: payload.Template, Ordering: payload.Ordering})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func DeleteRewrite(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		rewriteID, err := strconv.ParseInt(c.Param("rewriteId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := data.DeleteRewrite(app.DB, rewriteID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
