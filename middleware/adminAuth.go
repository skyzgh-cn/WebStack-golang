package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取cookie
		adminId, err := c.Cookie("admin_id")
		if err != nil || adminId == "" {
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
		c.Next()
	}
}
