package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/controllers/index"
)

func IndexRoutersInit(r *gin.Engine) {
	defaultRouters := r.Group("/")
	{
		defaultRouters.GET("/", index.IndexController{}.Index)

		defaultRouters.GET("/about", index.IndexController{}.About)

	}
}
