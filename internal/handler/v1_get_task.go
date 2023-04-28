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

type getTaskURI struct {
	TaskId int `uri:"taskId" binding:"required"`
}

func V1GetTask(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u getTaskURI
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

		task, err := postgres.GetTask(ctx, username, security.SaltPassword(password), u.TaskId)
		if err != nil {
			if errors.Is(err, model.ErrTaskNotFound) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Error:   typeTaskNotFound,
					Comment: strconv.Itoa(u.TaskId),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "get task",
			})

			return
		}

		c.JSON(http.StatusCreated, task)
	}
}
