package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type GroupController struct{}

func (gc GroupController) Index(c *gin.Context) {
	groups := []models.Group{}

	models.DB.Order("sort asc").Find(&groups)

	c.HTML(200, "admin/group.html", gin.H{
		"groups":    groups,
		"title":     "分类管理",
		"csrfToken": middleware.CSRFToken(c),
	})
	//c.JSON(200, gin.H{
	//	"groups": groups,
	//})
}

func (gc GroupController) Save(c *gin.Context) {

	var group models.Group
	if err := c.ShouldBind(&group); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 表单中主键字段名为 groupid（Group.Id 的 form 标签），0 表示新增
	if group.Id == 0 {
		// 新增分类：确保自增主键由数据库分配
		models.DB.Create(&group)
	} else {
		// 更新分类：按主键定位已有记录，只更新可变字段
		var existingGroup models.Group
		if err := models.DB.First(&existingGroup, group.Id).Error; err != nil {
			c.JSON(404, gin.H{"error": "分类不存在"})
			return
		}
		existingGroup.Name = group.Name
		existingGroup.Logo = group.Logo
		existingGroup.Sort = group.Sort
		models.DB.Save(&existingGroup)
	}

	c.Redirect(302, "/admin/groups")
}

func (gc GroupController) Delete(c *gin.Context) {
	id := c.PostForm("id")

	// 删除分类前先解除其下网站的关联，避免产生孤儿数据
	models.DB.Model(&models.Website{}).Where("group_id = ?", id).Update("group_id", 0)

	models.DB.Delete(&models.Group{}, id)
	c.JSON(200, gin.H{"code": 0, "msg": "删除成功"})
}
