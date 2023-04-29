package http

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	return w.ResponseWriter.Write(b)
}

func requestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}

		c.Writer = blw
		c.Next()

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
