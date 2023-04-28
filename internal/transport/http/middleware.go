package http

import (
	"bytes"
	"net/http"

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

func errorsHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}

		c.Writer = blw
		c.Next()

		if c.Writer.Status() >= http.StatusBadRequest {
			logger.Errorf("[%d][%s] %s", c.Writer.Status(), c.Request.Method, c.FullPath())
		}
	}
}

//func accessVerification(ctx context.Context, postgres model.Postgres) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		if !strings.HasPrefix(c.FullPath(), "/api") {
//			return
//		}
//
//		username, password, ok := c.Request.BasicAuth()
//		if !ok {
//			c.AbortWithStatus(http.StatusUnauthorized)
//
//			return
//		}
//
//		if err := postgres.CheckUser(ctx, username, security.SaltPassword(password)); err != nil {
//			if errors.Is(err, sql.ErrNoRows) {
//				c.AbortWithStatus(http.StatusForbidden)
//
//				return
//			}
//
//			c.JSON(http.StatusInternalServerError, handler.HTTPError{
//				Error:   "INTERNAL",
//				Comment: "check user",
//			})
//
//			return
//		}
//	}
//}
