package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
	"os"
	"path/filepath"
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

	// 绑定表单数据（不包含文件）
	if err := c.ShouldBind(&site); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取当前数据库中的站点信息以获取旧的 Logo 路径
	var existingSite models.Site
	models.DB.First(&existingSite, site.Id)

	// 获取上传的文件
	file, err := c.FormFile("sitelogo")
	if err == nil {

		ext := filepath.Ext(file.Filename)
		uploadDir := "./upload"
		dst := filepath.Join(uploadDir, "logo"+ext)

		// 创建上传目录（如果不存在）
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, os.ModePerm)
		}

		// 保存上传的文件（覆盖已有的 logo 文件）
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.HTML(500, "admin/settings.html", gin.H{
				"title": "系统设置",
				"site":  site,
				"error": "文件保存失败：" + err.Error(),
			})
			return
		}

		// 更新站点 Logo 路径
		site.Sitelogo = "/upload/logo" + ext // 假设前端访问路径为 /upload/
	} else {
		// 如果没有新上传文件，则保留旧的 Logo 路径
		site.Sitelogo = existingSite.Sitelogo
	}

	// 保存更新后的站点信息
	models.DB.Save(&site)

	// 重定向回设置页面
	c.Redirect(302, "/admin/settings")
}
