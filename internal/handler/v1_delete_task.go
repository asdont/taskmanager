package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

type deleteTaskURI struct {
	TaskID int `uri:"taskId" binding:"required"`
}

func V1DeleteTask(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u deleteTaskURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   typeParameterRequired,
				Comment: "task id",
			})

			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		if err := postgres.DeleteTask(ctx, username, security.SaltPassword(password), u.TaskID); err != nil {
			if errors.Is(err, model.ErrTaskNotFound) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Error:   typeTaskNotFound,
					Comment: strconv.Itoa(u.TaskID),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "delete task",
			})

			return
		}

		c.Status(http.StatusNoContent)
	}
}
