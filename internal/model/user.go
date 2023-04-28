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
	Id       int       `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
}

func (p Postgres) CreateNewUser(ctx context.Context, username string, password string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var userId int

	if err := p.Pool.QueryRowContext(ctx, `
		INSERT INTO
			auth(username, password, created)
		VALUES
		    ($1, $2, now())
		RETURNING
		    user_id
	`,
		username,
		password,
	).Scan(
		&userId,
	); err != nil {
		if db.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("%s: %w: %v", username, ErrUserAlreadyExists, err)
		}

		return 0, fmt.Errorf("query row: %w", err)
	}

	return userId, nil
}

func (p Postgres) CheckUser(ctx context.Context, username string, password string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var id int

	if err := p.Pool.QueryRowContext(ctx, `
		SELECT
		    user_id
		FROM
		    auth
		WHERE
		    username = $1 AND
		    password = $2
	`,
		username,
		password,
	).Scan(&id); err != nil {
		return fmt.Errorf("query row: %w", err)
	}

	return nil
}
