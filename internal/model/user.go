package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"taskmanager/internal/db"
)

type Postgres struct {
	Pool         *sql.DB
	Logger       *logrus.Logger
	QueryTimeout int
}

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
}

func (p Postgres) CreateNewUser(ctx context.Context, username string, password string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var userID int

	//nolint:execinquery
	if err := p.Pool.QueryRowContext(ctx, `
		INSERT INTO
			auth(username, password)
		VALUES
		    ($1, $2)
		RETURNING
		    user_id
	`,
		username,
		password,
	).Scan(
		&userID,
	); err != nil {
		if db.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("%s: %w: %s", username, ErrUserAlreadyExists, err.Error())
		}

		return 0, fmt.Errorf("query row: %w", err)
	}

	return userID, nil
}

func (p Postgres) DeleteUser(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	res, err := p.Pool.ExecContext(ctx, `
		DELETE FROM
		    auth
		WHERE
		    user_id = $1
	`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("userID %d: rows affected %d: %w", userID, rowsAffected, ErrUserNotFound)
	}

	return nil
}
