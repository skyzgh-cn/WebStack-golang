package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/auth"
)

// CSRFHeader 请求头名称，AJAX 请求通过它携带 CSRF 令牌。
const CSRFHeader = "X-CSRF-Token"

// CSRFField 表单字段名称，普通表单通过隐藏域携带 CSRF 令牌。
const CSRFField = "_csrf"

// CSRFToken 返回当前请求所属会话的 CSRF 令牌，未登录时返回空字符串。
func CSRFToken(c *gin.Context) string {
	if token, ok := c.Get("admin_csrf_token"); ok {
		if s, ok := token.(string); ok {
			return s
		}
	}
	return ""
}

// CSRF 校验所有写操作（POST/PUT/PATCH/DELETE），要求请求携带与会话匹配的令牌。
//
// 采用"会话内令牌 + 恒定时间比较"模型：令牌存储在服务端会话中，攻击者即使能
// 构造跨站请求也无法得知令牌值，因此无法通过校验。
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			c.Next()
			return
		}

		sess := auth.SessionFromRequest(c.Request)
		if sess == nil {
			// 未登录时交由认证中间件处理，这里直接拒绝写操作
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "请先登录",
			})
			return
		}

		token := c.GetHeader(CSRFHeader)
		if token == "" {
			token = c.PostForm(CSRFField)
		}
		if token == "" {
			token = c.Query(CSRFField)
		}

		// 使用恒定时间比较，避免通过响应耗时侧信道逐字节猜测令牌
		if subtle.ConstantTimeCompare([]byte(token), []byte(sess.CSRFToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"msg":  "CSRF 校验失败，请刷新页面后重试",
			})
			return
		}

		c.Set("admin_csrf_token", sess.CSRFToken)
		c.Next()
	}
}
