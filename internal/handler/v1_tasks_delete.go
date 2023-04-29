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

// V1DeleteTasks
//
// @Summary delete tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} deleteTasksResult
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/tasks [delete]
func V1DeleteTasks(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		quantity, err := postgres.DeleteTasks(ctx, username, security.SaltPassword(password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "delete task",
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, deleteTasksResult{
			Quantity: quantity,
		})
	}
}
