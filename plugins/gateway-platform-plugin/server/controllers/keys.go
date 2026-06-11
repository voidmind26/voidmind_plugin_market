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

type keyPayload struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

func keyFromPayload(payload keyPayload) models.Key {
	return models.Key{
		Name:        payload.Name,
		Value:       payload.Value,
		Description: payload.Description,
		Source:      payload.Source,
	}
}

func ListKeys(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := data.ListKeys(app.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func CreateKey(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload keyPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.CreateKey(app.DB, keyFromPayload(payload))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, item)
	}
}

func GetKey(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.GetKey(app.DB, id)
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

func UpdateKey(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var payload keyPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item, err := data.UpdateKey(app.DB, id, keyFromPayload(payload))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

func DeleteKey(app *helpers.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := data.DeleteKey(app.DB, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
