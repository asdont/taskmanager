package model

import (
	"context"
	"fmt"
	"time"

	"taskmanager/internal/db"
)

func (p Postgres) CreateTaskWithInjection(ctx context.Context, username, password, title string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var taskID int

	if err := p.Pool.QueryRowContext(ctx, fmt.Sprintf(`
			INSERT INTO
				task(user_id, status, title, created, updated)
			VALUES
    			((SELECT user_id FROM auth WHERE username = '%s' AND password = '%s'), %v, '%s', now(), now()) RETURNING  task_id			
		`,
		username,
		password,
		false,
		title,
	)).Scan(
		&taskID,
	); err != nil {
		if db.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("%s: %w: %s", username, ErrTaskAlreadyExists, err.Error())
		}

		if db.IsNullValueError(err) {
			return 0, fmt.Errorf("%s: %w: %s", username, ErrUserNotFound, err.Error())
		}

		return 0, fmt.Errorf("query row: %w", err)
	}

	return taskID, nil
}
