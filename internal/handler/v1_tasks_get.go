package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/security"
)

// V1GetTasks
//
// @Summary get tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} []model.Task
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/tasks [get]
func V1GetTasks(ctx context.Context, postgres PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		tasks, err := postgres.GetTasks(ctx, username, security.SaltPassword(password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "get tasks",
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, tasks)
	}
}
