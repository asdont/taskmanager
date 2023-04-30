package httpsrv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swagFiles "github.com/swaggo/files"
	swag "github.com/swaggo/gin-swagger"

	"taskmanager/internal/app"
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
	MaxShutdownTime    int
	CORS
}

type CORS struct {
	AllowHeaders []string
	AllowMethods []string
	AllowOrigins []string
}

func (conf Conf) RunHTTPServer(
	ctx context.Context, postgres model.Postgres, metrics app.Metrics, logger *logrus.Logger,
) error {
	gin.DisableConsoleColor()
	gin.SetMode(conf.Mode)
	gin.DefaultWriter = io.MultiWriter(os.Stdout)

	confCors := cors.DefaultConfig()
	confCors.AllowHeaders = conf.AllowHeaders
	confCors.AllowMethods = conf.AllowMethods
	confCors.AllowOrigins = conf.AllowOrigins

	router := gin.New()

	router.Use(
		gin.Recovery(),
		cors.New(confCors),
		requestLogger(metrics, logger),
		gin.LoggerWithFormatter(createLoggerFormatter()),
	)

	conf.setRouters(ctx, postgres, router, metrics)

	server := &http.Server{
		Addr:           ":" + conf.Port,
		Handler:        router,
		ReadTimeout:    time.Second * time.Duration(conf.ReadTimeoutSecond),
		WriteTimeout:   time.Second * time.Duration(conf.WriteTimeoutSecond),
		MaxHeaderBytes: conf.MaxHeaderBytes,
	}

	chErr := make(chan error)

	go func() {
		logger.Infof("start http server: port %s", conf.Port)

		if err := server.ListenAndServe(); err != nil {
			chErr <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	if err := conf.waitingStopSignal(ctx, server, chErr); err != nil {
		return fmt.Errorf("waiting server to stop: %w", err)
	}

	logger.Info("stop server: ok")

	return nil
}

func (conf Conf) setRouters(ctx context.Context, postgres handler.PostgresDB, router *gin.Engine, metrics app.Metrics) {
	// Swagger(OpenAPI).
	router.GET("/doc/*any", swag.WrapHandler(swagFiles.Handler))

	// Metrics(Prometheus).
	router.GET(metrics.MetricsRoute, gin.WrapH(promhttp.Handler()))

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

func createLoggerFormatter() func(p gin.LogFormatterParams) string {
	return func(p gin.LogFormatterParams) string {
		return fmt.Sprintf("%s | %15s | %s | %12s | %d | %6s | %s | %s | %s\n",
			p.TimeStamp.Format(time.RFC3339),
			p.ClientIP,
			p.Request.Proto,
			p.Latency,
			p.StatusCode,
			p.Method,
			p.Path,
			p.Request.UserAgent(),
			p.ErrorMessage,
		)
	}
}

//nolint:contextcheck
func (conf Conf) waitingStopSignal(ctx context.Context, server *http.Server, chErr chan error) error {
	select {
	case err := <-chErr:
		return err

	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.MaxShutdownTime))
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("forced stop: %w", err)
		}
	}

	return nil
}
