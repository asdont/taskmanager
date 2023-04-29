package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanager/internal/config"
	"taskmanager/internal/db"
	"taskmanager/internal/loggers"
	"taskmanager/internal/model"
	"taskmanager/internal/transport/http"

	_ "github.com/lib/pq"

	_ "taskmanager/docs"
)

// @title API Task Manager
// @version 1.0

// @contact.name Example
// @contact.url https://example.com/
// @contact.email example@mail.com

// @host 127.0.0.1:45222
// @schemes http
// @BasePath /api

// @securityDefinitions.basic BasicAuth
func main() {
	conf, err := config.Get("configs/conf.toml")
	if err != nil {
		log.Fatalf("get config: %v", err)
	}

	loggerConf := loggers.Conf{
		FileName:    conf.Logger.FileName,
		TimeFormat:  "06-01-02 15:04:05",
		MaxSizeMb:   conf.Logger.MaxSizeMb,
		MaxBackups:  conf.Logger.MaxBackups,
		MaxAgeDays:  conf.Logger.MaxAgeDays,
		MultiWriter: true, // stdout + file
		Level:       logrus.InfoLevel,
	}

	logger := loggerConf.CreateLoggerWithRotate("logs/app.log")

	postgresConf := db.Conf{
		EnvDockerConn: "DB_CONN",
		ConnAddress:   conf.Postgres.ConnAddress,
		MaxOpenConns:  conf.Postgres.MaxOpenConns,
		MaxIdleConns:  conf.Postgres.MaxIdleConns,
	}

	postgresPool, err := postgresConf.CreatePool(logger)
	if err != nil {
		logger.Fatalf("create postgres pool: %v", err)
	}

	postgres := model.Postgres{
		Pool:         postgresPool,
		QueryTimeout: conf.Postgres.QueryTimeout,
		Logger:       logger,
	}

	serverConf := http.Conf{
		Port:               conf.Server.Port,
		ManageUsername:     conf.Server.ManageUsername,
		ManagePassword:     conf.Server.ManagePassword,
		Mode:               gin.ReleaseMode,
		MaxHeaderBytes:     1 << 16, //nolint:gomnd
		ReadTimeoutSecond:  conf.Server.ReadTimeoutSeconds,
		WriteTimeoutSecond: conf.Server.WriteTimeoutSeconds,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := serverConf.RunHTTPServer(ctx, postgres, logger); err != nil {
		logger.Fatalf("run http server: %v", err)
	}
}
