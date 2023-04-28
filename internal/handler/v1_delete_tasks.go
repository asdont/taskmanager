package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

type deleteTasksResult struct {
	Quantity int64 `json:"quantity"`
}

func V1DeleteTasks(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		quantity, err := postgres.DeleteTasks(ctx, username, security.SaltPassword(password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "delete task",
			})

			return
		}

		c.JSON(http.StatusOK, deleteTasksResult{
			Quantity: quantity,
		})
	}
}
