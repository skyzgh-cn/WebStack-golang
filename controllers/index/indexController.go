package index

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type IndexController struct {
}

func (con IndexController) Index(c *gin.Context) {

	group := []models.Group{}
	models.DB.Preload("Websites").Order("sort asc").Find(&group)

	site := models.Site{}
	models.DB.First(&site)

	c.HTML(200, "index/index.html", gin.H{
		"groups":    group,
		"site":      site,
		"aboutweb":  template.HTML(site.Aboutweb),
		"aboutme":   template.HTML(site.Aboutme),
		"copyright": template.HTML(site.Copyright),
	})
	// c.JSON(200, gin.H{
	// 	"site": site,
	// })

}

func (con IndexController) About(c *gin.Context) {

	site := models.Site{}
	models.DB.First(&site)

	c.HTML(200, "index/about.html", gin.H{
		"site":      site,
		"aboutweb":  template.HTML(site.Aboutweb),
		"aboutme":   template.HTML(site.Aboutme),
		"copyright": template.HTML(site.Copyright),
	})
}
