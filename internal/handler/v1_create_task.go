package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

const maxLengthTaskTitle = 200

type createTaskBody struct {
	Title string `json:"title" binding:"required"`
}

type createTaskResult struct {
	TaskID int `json:"taskId"`
}

func V1CreateTask(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b createTaskBody
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   typeParameterRequired,
				Comment: "title",
			})

			return
		}

		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		if utf8.RuneCountInString(b.Title) > maxLengthTaskTitle {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   typeParameterTooLong,
				Comment: fmt.Sprintf("max %d", maxLengthTaskTitle),
			})

			return
		}

		taskID, err := postgres.CreateTask(ctx, username, security.SaltPassword(password), b.Title)
		if err != nil {
			if errors.Is(err, model.ErrUserNotFound) {
				c.AbortWithStatus(http.StatusUnauthorized)

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "create task",
			})

			return
		}

		c.JSON(http.StatusCreated, createTaskResult{
			TaskID: taskID,
		})
	}
}
