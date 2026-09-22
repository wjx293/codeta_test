package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// withTimeout 创建带超时的 context（基于请求 context）。
func withTimeout(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 10*time.Second)
}
