package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPError struct {
	Error   string `json:"-"`
	Type    string `json:"type,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// 401.
const (
	typeParameterTooLong      = "PARAMETER_TOO_LONG"
	typeParameterRequired     = "PARAMETER_REQUIRED"
	typeParametersRequired    = "PARAMETERS_REQUIRED"
	typePasswordRequired      = "PASSWORD_REQUIRED"
	typeTaskNotFound          = "TASK_NOT_FOUND"
	typeUsernameAlreadyExists = "USERNAME_ALREADY_EXISTS"
	typeUsernameRequired      = "USERNAME_REQUIRED"
	typeUserNotFound          = "USER_NOT_FOUND"
)

// 500.
const (
	typeInternalError = "INTERNAL"
)

func abortWithStatusUnauthorized(c *gin.Context) {
	c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	c.AbortWithStatus(http.StatusUnauthorized)
}
