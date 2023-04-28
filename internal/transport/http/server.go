package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanager/internal/handler"
	"taskmanager/internal/model"
)

type Conf struct {
	Port               string
	ManagementUsername string
	ManagementPassword string
	Mode               string
	MaxHeaderBytes     int
	ReadTimeoutSecond  int
	WriteTimeoutSecond int
}

func (conf Conf) RunHTTPServer(ctx context.Context, postgres model.Postgres, logger *logrus.Logger) error {
	loggerWriter := logger.Writer()

	defer func(loggerWriter *io.PipeWriter) {
		if err := loggerWriter.Close(); err != nil {
			logger.Errorf("logger writer: %v", err)
		}
	}(loggerWriter)

	gin.SetMode(conf.Mode)

	router := gin.New()

	router.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(loggerWriter),
		errorsHandler(logger),
	)

	conf.setRouters(ctx, postgres, router)

	server := &http.Server{
		Addr:           ":" + conf.Port,
		Handler:        router,
		ReadTimeout:    time.Second * time.Duration(conf.ReadTimeoutSecond),
		WriteTimeout:   time.Second * time.Duration(conf.WriteTimeoutSecond),
		MaxHeaderBytes: conf.MaxHeaderBytes,
	}

	logger.Infof("start http server: port %s", conf.Port)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (conf Conf) setRouters(ctx context.Context, postgres model.Postgres, router *gin.Engine) {
	management := router.Group(
		"/management/v1", gin.BasicAuth(gin.Accounts{
			conf.ManagementUsername: conf.ManagementPassword},
		),
	)
	{
		management.POST("/user", handler.V1CreateUser(ctx, postgres))
	}

	v1 := router.Group("/api/v1")
	{
		v1.POST("/task", handler.V1CreateTask(ctx, postgres))
		v1.GET("/task/:taskId", handler.V1GetTask(ctx, postgres))
		v1.PUT("/task/:taskId", handler.V1UpdateTask(ctx, postgres))
		v1.DELETE("/task/:taskId", handler.V1DeleteTask(ctx, postgres))

		v1.GET("/tasks", handler.V1GetTasks(ctx, postgres))
		v1.DELETE("/tasks", handler.V1DeleteTasks(ctx, postgres))
	}
}
