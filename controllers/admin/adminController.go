package admin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type AdminController struct{}

// Login 登录页面
func (con AdminController) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/login.html", gin.H{
		"title": "管理员登录",
	})
}

// HandleLogin 处理登录请求
func (con AdminController) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	admin := &models.Admin{}
	if err := admin.Login(username, password); err != nil {
		fmt.Printf("登录失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	// 设置session
	c.SetCookie("admin_id", admin.Username, 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "登录成功",
	})

}

// Dashboard 仪表盘
func (con AdminController) Dashboard(c *gin.Context) {
	// 获取系统信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取网站和分类统计
	var websiteCount, categoryCount int64
	models.DB.Model(&models.Website{}).Count(&websiteCount)
	models.DB.Model(&models.Group{}).Count(&categoryCount)

	// 获取最近添加的网站
	var recentWebsites []models.Website
	models.DB.Order("created_at desc").Limit(10).Find(&recentWebsites)

	// 获取每个分类下网站数量
	type GroupStat struct {
		Name  string
		Count int
	}
	var groupStats []GroupStat
	models.DB.Table("websites").
		Select("groups.name as name, count(websites.id) as count").
		Joins("left join groups on websites.group_id = groups.id").
		Group("groups.name").
		Scan(&groupStats)

	c.HTML(200, "admin/dashboard.html", gin.H{
		"title":          "管理后台",
		"websiteCount":   websiteCount,
		"categoryCount":  categoryCount,
		"recentWebsites": recentWebsites,
		"groupStats":     groupStats,
		"systemInfo": gin.H{
			"OS":          runtime.GOOS,
			"GoVersion":   runtime.Version(),
			"CPUCount":    runtime.NumCPU(),
			"MemoryUsage": m.Alloc / 1024 / 1024, // MB
		},
	})
}

// Logout 退出登录
func (con AdminController) Logout(c *gin.Context) {
	c.SetCookie("admin_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

// User 用户管理列表页面
func (con AdminController) User(c *gin.Context) {

	users := []models.Admin{}

	models.DB.Find(&users)

	c.HTML(http.StatusOK, "admin/user.html", gin.H{
		"title": "用户管理",
		"users": users,
	})
	//c.JSON(200, gin.H{"users": users})

}

// User 用户保存
func (con AdminController) Save(c *gin.Context) {
	ID := c.PostForm("id")
	var user models.Admin

	// 绑定表单数据到模型
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 如果是新增用户，或设置了新密码，则进行MD5加密
	if ID == "0" || user.Password != "" {
		h := md5.New()
		h.Write([]byte(user.Password))
		user.Password = hex.EncodeToString(h.Sum(nil))
	}

	if ID == "0" {
		// 新增用户
		models.DB.Create(&user)
	} else {
		// 更新用户（仅更新用户名和密码）
		var existingUser models.Admin
		models.DB.First(&existingUser, ID)
		existingUser.Username = user.Username
		if user.Password != "" {
			existingUser.Password = user.Password
		}
		models.DB.Save(&existingUser)
	}

	c.Redirect(302, "/admin/user")
}

// User 用户删除
func (con AdminController) Delete(c *gin.Context) {
	id := c.PostForm("id")

	// 获取要删除的用户
	var user models.Admin
	if err := models.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "用户不存在",
		})
		return
	}

	// 检查是否是最后一个用户
	var count int64
	models.DB.Model(&models.Admin{}).Count(&count)
	if count <= 1 {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "不能删除最后一个用户",
		})
		return
	}

	// 执行删除操作
	models.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "删除成功",
	})
}
