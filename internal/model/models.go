package model

import (
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrTaskNotFound      = errors.New("task not found")
)
