package db

import (
	"errors"

	"github.com/lib/pq"
)

const (
	postgresUniqueConstraintError = "23505"
	postgresNullValueError        = "23502"
)

func IsUniqueConstraintError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) {
		if pqError.Code == postgresUniqueConstraintError {
			return true
		}
	}

	return false
}

func IsNullValueError(err error) bool {
	var pqError *pq.Error
	if errors.As(err, &pqError) {
		if pqError.Code == postgresNullValueError {
			return true
		}
	}

	return false
}
