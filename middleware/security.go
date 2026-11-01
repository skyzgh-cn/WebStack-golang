package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有响应添加基础安全响应头，降低点击劫持、内容嗅探等风险。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-XSS-Protection", "0")
		c.Next()
	}
}

// MaxBodySize 限制请求体大小，防止超大请求耗尽内存或磁盘。
func MaxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
