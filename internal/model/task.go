package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"taskmanager/internal/db"
)

type Task struct {
	ID        int       `json:"id,omitempty"`
	Status    bool      `json:"status"`
	Title     string    `json:"title"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Completed time.Time `json:"completed"`
}

func (p Postgres) CreateTask(ctx context.Context, username, password, title string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var taskID int

	//nolint:execinquery
	if err := p.Pool.QueryRowContext(ctx, `
		INSERT INTO
			task(user_id, status, title, created, updated)
		VALUES
		    ((SELECT user_id FROM auth WHERE username = $1 AND password = $2), $3, $4, now(), now())
		RETURNING
		    task_id
	`,
		username,
		password,
		false,
		title,
	).Scan(
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

func (p Postgres) GetTask(ctx context.Context, username, password string, taskID int) (Task, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	var (
		task          Task
		taskCompleted sql.NullTime
	)

	if err := p.Pool.QueryRowContext(ctx, `
		SELECT
		    t.status, t.title, t.created, t.updated, t.completed
		FROM
		    task t 
		JOIN
		    auth a USING (user_id)
		WHERE
		    a.username = $1 AND
		    a.password = $2 AND
		    t.task_id = $3
	`,
		username,
		password,
		taskID,
	).Scan(
		&task.Status,
		&task.Title,
		&task.Created,
		&task.Updated,
		&taskCompleted,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, fmt.Errorf("taskId: %d: %w", taskID, ErrTaskNotFound)
		}

		return Task{}, fmt.Errorf("query row: %w", err)
	}

	task.Completed = taskCompleted.Time

	return task, nil
}

func (p Postgres) GetTasks(ctx context.Context, username, password string) ([]Task, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	rows, err := p.Pool.QueryContext(ctx, `
		SELECT
		    t.task_id, t.status, t.title, t.created, t.updated, t.completed
		FROM
		    task t 
		JOIN
		    auth a USING (user_id)
		WHERE
		    a.username = $1 AND
		    a.password = $2
	`,
		username,
		password,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			p.Logger.Errorf("get tasks: %v", err)
		}
	}()

	var tasks []Task

	for rows.Next() {
		var (
			task          Task
			taskCompleted sql.NullTime
		)

		if err := rows.Scan(
			&task.ID,
			&task.Status,
			&task.Title,
			&task.Created,
			&task.Updated,
			&taskCompleted,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		task.Completed = taskCompleted.Time

		tasks = append(tasks, task)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("scan rows: %w", rows.Err())
	}

	return tasks, nil
}

func (p Postgres) UpdateTask(ctx context.Context, username, password string, taskID int, setValues []string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	res, err := p.Pool.ExecContext(ctx, fmt.Sprintf(`
		UPDATE
			task
		SET
		   %s
		WHERE
		    user_id = (SELECT user_id FROM auth WHERE username = %s AND password = %s) AND
		    task_id = %d
	`,
		strings.Join(setValues, ","),
		pq.QuoteLiteral(username),
		pq.QuoteLiteral(password),
		taskID,
	))
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("taskId %d: rows affected %d: %w", taskID, rowsAffected, ErrTaskNotFound)
	}

	return nil
}

func (p Postgres) DeleteTask(ctx context.Context, username, password string, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	res, err := p.Pool.ExecContext(ctx, `
		DELETE FROM
		    task
		WHERE
		    user_id = (SELECT user_id FROM auth WHERE username = $1 AND password = $2) AND
		    task_id = $3
	`,
		username,
		password,
		taskID,
	)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("taskId %d: rows affected %d: %w", taskID, rowsAffected, ErrTaskNotFound)
	}

	return nil
}

func (p Postgres) DeleteTasks(ctx context.Context, username, password string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(p.QueryTimeout))
	defer cancel()

	res, err := p.Pool.ExecContext(ctx, `
		DELETE FROM
		    task
		WHERE
		    user_id = (SELECT user_id FROM auth WHERE username = $1 AND password = $2)
	`,
		username,
		password,
	)
	if err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}

	return rowsAffected, nil
}
