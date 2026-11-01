package admin

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/auth"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type AdminController struct{}

// isSecureRequest 判断当前请求是否经由 HTTPS，用于决定 Cookie 的 Secure 标记。
//
// 若配置中显式启用了 secureCookie，则始终视为安全连接；
// 否则依据 TLS 与反向代理头判断。由于 Gin 只信任配置中列出的代理，
// 客户端无法通过伪造 X-Forwarded-Proto 头影响该判断。
func isSecureRequest(c *gin.Context) bool {
	if config, err := models.LoadConfig(); err == nil && config.App.SecureCookie {
		return true
	}
	if c.Request.TLS != nil {
		return true
	}
	return c.GetHeader("X-Forwarded-Proto") == "https"
}

// Login 登录页面
func (con AdminController) Login(c *gin.Context) {
	// 签发登录 CSRF 令牌（双提交 Cookie 模式）
	csrfToken, err := auth.IssueLoginCSRF(c.Writer, isSecureRequest(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "生成令牌失败",
		})
		return
	}

	c.HTML(http.StatusOK, "admin/login.html", gin.H{
		"title":     "管理员登录",
		"csrfToken": csrfToken,
	})
}

// HandleLogin 处理登录请求
func (con AdminController) HandleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 校验登录 CSRF 令牌，阻断跨站登录请求伪造
	if !auth.VerifyLoginCSRF(c.Request, c.PostForm(auth.LoginCSRFField)) {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "CSRF 校验失败，请刷新页面后重试",
		})
		return
	}

	// 限制登录失败次数，防止暴力破解
	clientIP := c.ClientIP()
	if !auth.AllowLogin(clientIP, username) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"code": 1,
			"msg":  "登录尝试过于频繁，请稍后再试",
		})
		return
	}

	admin := &models.Admin{}
	if err := admin.Login(username, password); err != nil {
		auth.RecordLoginFailure(clientIP, username)
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  err.Error(),
		})
		return
	}

	auth.ResetLoginFailures(clientIP, username)

	// 创建服务端会话，Cookie 中仅保存随机会话令牌
	token, _, err := auth.Default().Create(admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "创建会话失败",
		})
		return
	}

	auth.SetSessionCookie(c.Writer, token, isSecureRequest(c))

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
		Select("`groups`.`name` as name, COUNT(websites.id) as count").
		Joins("RIGHT JOIN `groups` ON websites.group_id = `groups`.`id`").
		Group("`groups`.`name`").
		Scan(&groupStats)

	c.HTML(200, "admin/dashboard.html", gin.H{
		"title":          "管理后台",
		"websiteCount":   websiteCount,
		"categoryCount":  categoryCount,
		"recentWebsites": recentWebsites,
		"groupStats":     groupStats,
		"csrfToken":      middleware.CSRFToken(c),
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
	if cookie, err := c.Request.Cookie(auth.CookieName); err == nil {
		auth.Default().Delete(cookie.Value)
	}
	auth.ClearSessionCookie(c.Writer, isSecureRequest(c))
	c.Redirect(http.StatusFound, "/admin/login")
}

// User 用户管理列表页面
func (con AdminController) User(c *gin.Context) {

	users := []models.Admin{}

	models.DB.Find(&users)

	c.HTML(http.StatusOK, "admin/user.html", gin.H{
		"title":     "用户管理",
		"users":     users,
		"csrfToken": middleware.CSRFToken(c),
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

	// 如果是新增用户，或设置了新密码，则使用 bcrypt 加密
	if ID == "0" || user.Password != "" {
		hashed, err := models.HashPassword(user.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": "密码加密失败"})
			return
		}
		user.Password = hashed
	}

	if ID == "0" {
		// 新增用户
		models.DB.Create(&user)
	} else {
		// 更新用户（仅更新用户名和密码）
		var existingUser models.Admin
		if err := models.DB.First(&existingUser, ID).Error; err != nil {
			c.JSON(404, gin.H{"error": "用户不存在"})
			return
		}
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
