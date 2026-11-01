package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/auth"
)

// AdminAuth 校验后台请求是否携带有效的登录会话。
//
// 安全要点：不再仅凭 admin_id Cookie 是否存在来放行，而是校验服务端会话
// 令牌是否真实存在、未过期。未登录时普通请求跳转登录页，AJAX 请求返回 401。
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := auth.SessionFromRequest(c.Request)
		if sess == nil {
			// 如果是AJAX请求
			if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "请先登录",
				})
				c.Abort()
				return
			}
			// 普通请求重定向到登录页
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		c.Set("admin_username", sess.Username)
		c.Set("admin_csrf_token", sess.CSRFToken)
		c.Next()
	}
}
