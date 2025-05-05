package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/controllers/admin"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
)

func AdminRoutersInit(r *gin.Engine) {
	adminRouters := r.Group("/admin")
	{
		// 无需认证的路由
		adminRouters.GET("/login", admin.AdminController{}.Login)
		adminRouters.POST("/login", admin.AdminController{}.HandleLogin)

		// 需要认证的路由
		authorized := adminRouters.Group("/", middleware.AdminAuth())
		{
			// 修改：确保路由映射正确
			authorized.GET("/", admin.AdminController{}.Dashboard)
			authorized.GET("/dashboard", admin.AdminController{}.Dashboard)
			authorized.GET("/logout", admin.AdminController{}.Logout)
			authorized.GET("/user", admin.AdminController{}.User)
			authorized.POST("/user/save", admin.AdminController{}.Save)
			authorized.POST("/user/delete", admin.AdminController{}.Delete)

			// 新增网站设置路由
			authorized.GET("/settings", admin.SiteController{}.Settings)
			authorized.POST("/settings", admin.SiteController{}.UpdateSettings)

			// 新增分类管理路由
			authorized.GET("/groups", admin.GroupController{}.Index)
			authorized.POST("/groups/save", admin.GroupController{}.Save)
			authorized.POST("/groups/delete", admin.GroupController{}.Delete)

			//新增网站管理路由
			authorized.GET("/websites", admin.WebsiteController{}.Index)
			authorized.POST("/websites/save", admin.WebsiteController{}.Save)
			authorized.POST("/websites/delete", admin.WebsiteController{}.Delete)
			authorized.POST("/websites/fetch-meta", admin.WebsiteController{}.FetchMeta)

		}
	}
}
