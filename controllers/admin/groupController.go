package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type GroupController struct{}

func (gc GroupController) Index(c *gin.Context) {
	groups := []models.Group{}

	models.DB.Order("sort asc").Find(&groups)

	c.HTML(200, "admin/group.html", gin.H{
		"groups": groups,
		"title":  "分类管理",
	})
	//c.JSON(200, gin.H{
	//	"groups": groups,
	//})
}

func (gc GroupController) Save(c *gin.Context) {

	ID := c.PostForm("id")
	var group models.Group
	if ID == "0" {

		if err := c.ShouldBind(&group); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		models.DB.Create(&group)

	} else {

		if err := c.ShouldBind(&group); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		models.DB.Save(&group)
	}

	//c.JSON(200, gin.H{
	//	"group": group,
	//})

	c.Redirect(302, "/admin/groups")
}

func (gc GroupController) Delete(c *gin.Context) {
	id := c.PostForm("id")
	models.DB.Delete(&models.Group{}, id)
	c.JSON(200, gin.H{"code": 0, "msg": "删除成功"})
}
