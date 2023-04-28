package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Environment variable for Dockerfile.
const dockerEnvConnAddress = "DB_CONN"

const (
	maxConnectionAttempts = 5
	delayBetweenAttempts  = 5
)

var (
	errFailedConnection = errors.New("failed connection")
	errTempError        = errors.New("temp error")
)

type Conf struct {
	EnvDockerConn string
	ConnAddress   string
	MaxOpenConns  int
	MaxIdleConns  int
}

func (conf Conf) CreatePool(logger *logrus.Logger) (*sql.DB, error) {
	envDockerConnAddress, ok := os.LookupEnv(dockerEnvConnAddress)
	if ok {
		conf.ConnAddress = envDockerConnAddress
	}

	for i := 0; i <= maxConnectionAttempts; i++ {
		pool, err := createPool(conf.ConnAddress)
		if err != nil {
			if errors.Is(err, errTempError) {
				logger.Warnf("connection attempt %d: %v\n", i+1, err)

				time.Sleep(time.Second * delayBetweenAttempts)

				continue
			}

			return nil, fmt.Errorf("create pool: %w", err)
		}

		pool.SetMaxOpenConns(conf.MaxOpenConns)
		pool.SetMaxIdleConns(conf.MaxIdleConns)

		return pool, nil
	}

	return nil, errFailedConnection
}

func createPool(connAddress string) (*sql.DB, error) {
	pool, err := sql.Open("postgres", connAddress)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	if err = pool.Ping(); err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			if err := pool.Close(); err != nil {
				return nil, fmt.Errorf("pool: close: %w", err)
			}

			return nil, fmt.Errorf("%v: %w", err, errTempError)
		}

		return nil, fmt.Errorf("ping: fail: %w", err)
	}

	return pool, nil
}
