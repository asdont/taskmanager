package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
)

type deleteUserURI struct {
	UserID int `uri:"userId" binding:"required" example:"7"`
}

// V1DeleteUser here authorization is checked at the server level.
//
// @Summary delete user(manage auth - admin:admin)
// @Tags management
// @Accept json
// @Produce json
// @Param userId path int true "userId" minimum(1)
// @Success 204 {object} nil
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/manage/user/{userId} [delete]
func V1DeleteUser(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u deleteUserURI
		if err := c.ShouldBindUri(&u); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParameterRequired,
				Comment: "userId",
				Error:   err.Error(),
			})

			return
		}

		if err := postgres.DeleteUser(ctx, u.UserID); err != nil {
			if errors.Is(err, model.ErrUserNotFound) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Type:    typeUserNotFound,
					Comment: "userId",
					Error:   err.Error(),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "delete user",
				Error:   err.Error(),
			})

			return
		}

		c.Status(http.StatusNoContent)
	}
}
