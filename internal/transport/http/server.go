package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swagFiles "github.com/swaggo/files"
	swag "github.com/swaggo/gin-swagger"

	"taskmanager/internal/handler"
	"taskmanager/internal/model"
)

type Conf struct {
	Port               string
	ManageUsername     string
	ManagePassword     string
	Mode               string
	MaxHeaderBytes     int
	ReadTimeoutSecond  int
	WriteTimeoutSecond int
}

func (conf Conf) RunHTTPServer(ctx context.Context, postgres model.Postgres, logger *logrus.Logger) error {
	//gin.DisableConsoleColor()
	gin.SetMode(conf.Mode)
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

	router := gin.New()

	router.Use(
		gin.Recovery(),
		requestLogger(logger),
		gin.LoggerWithWriter(gin.DefaultWriter),
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
	router.GET("/doc/*any", swag.WrapHandler(swagFiles.Handler))

	api := router.Group("/api")
	v1 := api.Group("/v1")

	manage := v1.Group("/manage", gin.BasicAuth(gin.Accounts{conf.ManageUsername: conf.ManagePassword}))
	{
		manage.POST("/user", handler.V1CreateUser(ctx, postgres))
		manage.DELETE("/user/:userId", handler.V1DeleteUser(ctx, postgres))
	}

	task := v1.Group("/task")
	{
		task.POST("/", handler.V1CreateTask(ctx, postgres))
		task.GET("/:taskId", handler.V1GetTask(ctx, postgres))
		task.PUT("/:taskId", handler.V1UpdateTask(ctx, postgres))
		task.DELETE("/:taskId", handler.V1DeleteTask(ctx, postgres))
	}

	tasks := v1.Group("/tasks")
	{
		tasks.GET("/", handler.V1GetTasks(ctx, postgres))
		tasks.DELETE("/", handler.V1DeleteTasks(ctx, postgres))
	}
}
