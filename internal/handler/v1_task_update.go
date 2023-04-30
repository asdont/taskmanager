package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

type updateTaskURI struct {
	TaskID int `uri:"taskId" binding:"required" example:"24"`
}

type updateTaskBody struct {
	Title     string `json:"title" example:"some new title"`
	Completed bool   `json:"completed" example:"true"`
}

// V1UpdateTask
//
// @Summary update task
// @Tags task
// @Accept json
// @Produce json
// @Param taskId path int true "taskId" minimum(1)
// @Param data body updateTaskBody true "any of the fields"
// @Success 204 {object} nil
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/task/{taskId} [put]
func V1UpdateTask(ctx context.Context, postgres PostgresDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			abortWithStatusUnauthorized(c)

			return
		}

		var u updateTaskURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterRequired,
				Comment: "taskId",
				Error:   err.Error(),
			})

			return
		}

		var b updateTaskBody
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:  typeParametersRequired,
				Error: err.Error(),
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

		if err := postgres.UpdateTask(
			ctx, username, security.SaltPassword(password), u.TaskID, updateTaskCreateSetValues(b),
		); err != nil {
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
				Comment: "update task",
				Error:   err.Error(),
			})

			return
		}

		c.Status(http.StatusNoContent)
	}
}

func updateTaskCreateSetValues(b updateTaskBody) []string {
	setValues := []string{
		"status=" + strconv.FormatBool(b.Completed),
	}

	if b.Title != "" {
		setValues = append(setValues, "title="+pq.QuoteLiteral(b.Title))
	}

	if b.Completed {
		setValues = append(setValues, "completed=now()")
	} else {
		setValues = append(setValues, "updated=now()")
		setValues = append(setValues, "completed=null")
	}

	return setValues
}
