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
	TaskID int `uri:"taskId" binding:"required" example:"24"`
}

// V1DeleteTask
//
// @Summary delete task
// @Tags task
// @Accept json
// @Produce json
// @Param taskId path int true "taskId" minimum(1)
// @Success 204 {object} nil
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/task/{taskId} [delete]
func V1DeleteTask(ctx context.Context, postgres PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		var u deleteTaskURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterRequired,
				Comment: "taskId",
				Error:   err.Error(),
			})

			return
		}

		if err := postgres.DeleteTask(ctx, username, security.SaltPassword(password), u.TaskID); err != nil {
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
				Comment: "delete task",
				Error:   err.Error(),
			})

			return
		}

		c.Status(http.StatusNoContent)
	}
}
