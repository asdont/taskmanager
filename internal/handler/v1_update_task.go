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
	TaskID int `uri:"taskId" binding:"required"`
}

type updateTaskBody struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func V1UpdateTask(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u updateTaskURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error:   typeParameterRequired,
				Comment: "task id",
			})

			return
		}

		var b updateTaskBody
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Error: typeParametersRequired,
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

		if err := postgres.UpdateTask(
			ctx, username, security.SaltPassword(password), u.TaskID, updateTaskCreateSetValues(b),
		); err != nil {
			if errors.Is(err, model.ErrTaskNotFound) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Error:   typeTaskNotFound,
					Comment: strconv.Itoa(u.TaskID),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Error:   typeInternalError,
				Comment: "update task",
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
