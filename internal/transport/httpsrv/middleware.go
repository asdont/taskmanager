package httpsrv

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"taskmanager/internal/app"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func requestLogger(metrics app.Metrics, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}

		c.Writer = blw
		c.Next()

		// Ignore prometheus path.
		if c.FullPath() != metrics.MetricsRoute {
			metrics.RequestsTotal.WithLabelValues(
				strconv.Itoa(c.Writer.Status()),
				c.Request.Method,
				c.FullPath(),
			).Inc()
		}

		if c.Writer.Status() >= http.StatusBadRequest {
			logger.Errorf(
				"[SERVER] %5d | %15s | %6s | %s | %v",
				c.Writer.Status(),
				c.ClientIP(),
				c.Request.Method,
				c.FullPath(),
				strings.ReplaceAll(blw.body.String(), "\"", ""),
			)
		}
	}
}
