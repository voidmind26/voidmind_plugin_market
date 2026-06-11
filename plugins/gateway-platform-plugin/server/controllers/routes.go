package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"gateway-platform-plugin/server/data"
	"gateway-platform-plugin/server/helpers"
	"gateway-platform-plugin/server/models"

	"github.com/gin-gonic/gin"
)

type routePayload struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	LocalPath   string `json:"local_path"`
	UpstreamURL string `json:"upstream_url"`
	TimeoutMS   int    `json:"timeout_ms"`
	Description string `json:"description"`
}

func routeFromPayload(payload routePayload) models.Route {
	return models.Route{
		Name:        payload.Name,
		Enabled:     payload.Enabled,
		LocalPath:   payload.LocalPath,
		UpstreamURL: payload.UpstreamURL,
		TimeoutMS:   payload.TimeoutMS,
		Description: payload.Description,
	}
}

func ListRoutes(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := data.ListRoutes(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func CreateRoute(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload routePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.CreateRoute(app.DB, routeFromPayload(payload))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	}
}

func GetRoute(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.GetRoute(app.DB, id)
		if err != nil {
			status := http.StatusInternalServerError
			if err == sql.ErrNoRows {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func UpdateRoute(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var payload routePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.UpdateRoute(app.DB, id, routeFromPayload(payload))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func DeleteRoute(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := data.DeleteRoute(app.DB, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
