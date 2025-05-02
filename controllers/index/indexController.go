package index

import "github.com/gin-gonic/gin"

type IndexController struct {
}

func (con IndexController) Index(c *gin.Context) {
	c.HTML(200, "default/index.html", gin.H{})
}

func (con IndexController) About(c *gin.Context) {
	c.HTML(200, "default/about.html", gin.H{})
}
