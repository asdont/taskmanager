package model

import (
	"database/sql"
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrTaskNotFound      = errors.New("task not found")
)

type Postgres struct {
	Pool         *sql.DB
	Logger       *logrus.Logger
	QueryTimeout int
}
