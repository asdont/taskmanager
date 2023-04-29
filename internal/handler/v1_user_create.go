package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

const (
	minLengthUsername = 3
	maxLengthUsername = 20

	minLengthPassword = 5
	maxLengthPassword = 20
)

type createUserBody struct {
	Username string `json:"username" binding:"required" example:"newuser"`
	Password string `json:"password" binding:"required" example:"newuser"`
}

type createUserResult struct {
	UserID int `json:"userId"`
}

// V1CreateUser here authorization is checked at the server level.
//
// @Summary create new user(manage auth - admin:admin)
// @Tags management
// @Accept json
// @Produce json
// @Param data body createUserBody true "manage auth - admin:admin; username(3-20); password(5-20)"
// @Success 201 {object} createUserResult "userId"
// @Failure 400 {object} HTTPError "error type, comment"
// @Failure 401 {object} nil
// @Failure 500 {object} HTTPError "error type, comment"
// @Router /v1/manage/user [post]
func V1CreateUser(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b createUserBody
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, HTTPError{
				Type:    typeParametersRequired,
				Comment: "username and password required",
				Error:   err.Error(),
			})

			return
		}

		res := checkUsernamePassword(b.Username, b.Password)
		if res != nil {
			c.JSON(http.StatusBadRequest, res)

			return
		}

		userID, err := postgres.CreateNewUser(ctx, b.Username, security.SaltPassword(b.Password))
		if err != nil {
			if errors.Is(err, model.ErrUserAlreadyExists) {
				c.JSON(http.StatusBadRequest, HTTPError{
					Type:    typeUsernameAlreadyExists,
					Comment: b.Username,
					Error:   err.Error(),
				})

				return
			}

			c.JSON(http.StatusInternalServerError, HTTPError{
				Type:    typeInternalError,
				Comment: "create new user",
				Error:   err.Error(),
			})

			return
		}

		c.JSON(http.StatusCreated, createUserResult{
			UserID: userID,
		})
	}
}

func checkUsernamePassword(username, password string) *HTTPError {
	res := checkUsername(username)

	if res != nil {
		return res
	}

	res = checkPassword(password)

	if res != nil {
		return res
	}

	return nil
}

func checkUsername(username string) *HTTPError {
	count := 0

	for _, c := range username {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			count++

			continue
		}

		return &HTTPError{
			Type:    typeUsernameRequired,
			Comment: "only letters and numbers required",
		}
	}

	if count < minLengthUsername || count > maxLengthUsername {
		return &HTTPError{
			Comment: typeUsernameRequired,
			Type:    fmt.Sprintf("min %d, max %d", minLengthUsername, maxLengthUsername),
		}
	}

	return nil
}

func checkPassword(password string) *HTTPError {
	allowedSpecialChars := "!@#$%^&*"

	count := 0

	for _, c := range password {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || strings.Contains(allowedSpecialChars, string(c)) {
			count++

			continue
		}

		return &HTTPError{
			Type:    typePasswordRequired,
			Comment: fmt.Sprintf("only letters, numbers or %s required", allowedSpecialChars),
		}
	}

	if count < minLengthPassword || count > maxLengthPassword {
		return &HTTPError{
			Type:    typePasswordRequired,
			Comment: fmt.Sprintf("min %d, max %d", minLengthPassword, maxLengthPassword),
		}
	}

	return nil
}
