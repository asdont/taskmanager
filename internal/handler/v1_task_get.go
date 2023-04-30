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
	TaskID int `uri:"taskId" binding:"required" example:"24"`
}

// V1GetTask
//
// @Summary get task
// @Tags task
// @Accept json
// @Produce json
// @Param taskId path int true "taskId" minimum(1)
// @Success 200 {object} model.Task
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/task/{taskId} [get]
func V1GetTask(ctx context.Context, postgres PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		var u getTaskURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterRequired,
				Comment: "taskId",
				Error:   err.Error(),
			})

			return
		}

		task, err := postgres.GetTask(ctx, username, security.SaltPassword(password), u.TaskID)
		if err != nil {
			if errors.Is(err, model.ErrTaskNotFound) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Type:    typeTaskNotFound,
					Comment: strconv.Itoa(u.TaskID),
					Error:   err.Error(),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "get task",
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, task)
	}
}
