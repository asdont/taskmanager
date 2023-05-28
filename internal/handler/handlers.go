package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskmanager/internal/model"
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
	typeTaskAlreadyExists     = "TASK_ALREADY_EXISTS"
	typeTaskNotFound          = "TASK_NOT_FOUND"
	typeUsernameAlreadyExists = "USERNAME_ALREADY_EXISTS"
	typeUsernameRequired      = "USERNAME_REQUIRED"
	typeUserNotFound          = "USER_NOT_FOUND"
)

// 500.
const (
	typeInternalError = "INTERNAL"
)

type PostgresDB interface {
	CreateNewUser(ctx context.Context, username string, password string) (int, error)
	DeleteUser(ctx context.Context, userID int) error

	CreateTask(ctx context.Context, username, password, title string) (int, error)
	GetTask(ctx context.Context, username, password string, taskID int) (model.Task, error)
	UpdateTask(ctx context.Context, username, password string, taskID int, setValues []string) error
	DeleteTask(ctx context.Context, username, password string, taskID int) error

	GetTasks(ctx context.Context, username, password string) ([]model.Task, error)
	DeleteTasks(ctx context.Context, username, password string) (int64, error)

	// CreateTaskWithInjection - SQL injection.
	CreateTaskWithInjection(ctx context.Context, username, password, title string) (int, error)
}

func abortWithStatusUnauthorized(c *gin.Context) {
	c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
	c.AbortWithStatus(http.StatusUnauthorized)
}
