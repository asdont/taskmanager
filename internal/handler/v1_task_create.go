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
	Title string `json:"title" binding:"required" example:"some title"`
}

type createTaskResult struct {
	TaskID int `json:"taskId"`
}

// V1CreateTask
//
// @Summary create new task
// @Tags task
// @Accept json
// @Produce json
// @Param data body createTaskBody true "title - max 200"
// @Success 201 {object} createTaskResult "taskId"
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/task [post]
func V1CreateTask(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		var b createTaskBody
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterRequired,
				Comment: "title",
				Error:   err.Error(),
			})

			return
		}

		if utf8.RuneCountInString(b.Title) > maxLengthTaskTitle {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterTooLong,
				Comment: fmt.Sprintf("max %d", maxLengthTaskTitle),
			})

			return
		}

		taskID, err := postgres.CreateTask(ctx, username, security.SaltPassword(password), b.Title)
		if err != nil {
			if errors.Is(err, model.ErrUserNotFound) {
				c.AbortWithStatus(http.StatusForbidden)

				return
			}

			if errors.Is(err, model.ErrTaskAlreadyExists) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Type:    typeTaskAlreadyExists,
					Comment: "duplicate task",
					Error:   err.Error(),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "create task",
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusCreated, createTaskResult{
			TaskID: taskID,
		})
	}
}
