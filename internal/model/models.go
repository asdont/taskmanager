package model

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	ErrTaskNotFound = errors.New("task not found")
)
