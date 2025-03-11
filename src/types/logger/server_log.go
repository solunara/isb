package logger

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AccessLog struct {
	Method       string
	Url          string
	MatchedRoute string
	Body         string
}

type logMiddlewareBuilder struct {
	logger *zap.Logger
	input  bool
}

func NewMiddlewareLogBuilder(logger *zap.Logger) *logMiddlewareBuilder {
	return &logMiddlewareBuilder{
		logger: logger,
		input:  false,
	}
}

func (l *logMiddlewareBuilder) LogInput() *logMiddlewareBuilder {
	l.input = true
	return l
}

func (l logMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.String()
		if len(url) > 4096 {
			url = url[:4096]
		}
		al := AccessLog{
			Method: c.Request.Method,
			Url:    url,
		}
		if l.input {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			al.Body = string(body)
		}
		defer func() {
			al.MatchedRoute = c.FullPath()
			l.logger.Info("recv:", zap.Any("req", al))
		}()
		c.Next()
	}
}
