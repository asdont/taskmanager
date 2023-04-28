package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

func V1GetTasks(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		tasks, err := postgres.GetTasks(ctx, username, security.SaltPassword(password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "get tasks",
			})

			return
		}

		c.JSON(http.StatusCreated, tasks)
	}
}
