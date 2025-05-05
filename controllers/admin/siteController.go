package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type SiteController struct{}

// Settings 处理网站设置页面的 GET 请求
func (sc SiteController) Settings(c *gin.Context) {
	site := models.Site{}
	models.DB.First(&site)
	// 从数据库中读取网站设置

	c.HTML(200, "admin/settings.html", gin.H{
		"title": "系统设置",
		"site":  site,
	})
}

// UpdateSettings 处理网站设置页面的 POST 请求
func (sc SiteController) UpdateSettings(c *gin.Context) {
	var site models.Site
	if err := c.ShouldBind(&site); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 修改数据库中的网站设置
	models.DB.Save(&site)

	c.Redirect(302, "/admin/settings")
}
