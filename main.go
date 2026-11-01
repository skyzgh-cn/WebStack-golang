package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/auth"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
	"github.com/skyzgh-cn/WebStack-golang/models"
	"github.com/skyzgh-cn/WebStack-golang/routers"
	"github.com/skyzgh-cn/WebStack-golang/templatefuncs"
)

//go:embed templates/**/*
var templates embed.FS

//go:embed assets/*
var assets embed.FS

func main() {
	// 初始化数据库连接与表结构
	if err := models.InitDB(); err != nil {
		fmt.Println("初始化数据库失败:", err)
		return
	}

	config, err := models.LoadConfig() // 加载配置文件
	if err != nil {
		fmt.Println("无法加载配置文件:", err) // 如果加载失败，打印错误信息
		return
	}

	// 定期清理过期会话，避免内存中残留已失效的会话数据
	stopCleanup := make(chan struct{})
	defer close(stopCleanup)
	auth.Default().StartCleanupLoop(stopCleanup)

	r := gin.Default() // 创建 Gin 引擎实例

	// 只信任配置中列出的反向代理，避免客户端伪造 X-Forwarded-* 头
	if len(config.App.TrustedProxies) > 0 {
		if err := r.SetTrustedProxies(config.App.TrustedProxies); err != nil {
			fmt.Println("设置可信代理失败:", err)
		}
	} else {
		r.SetTrustedProxies(nil)
	}

	// 基础安全响应头
	r.Use(middleware.SecurityHeaders())

	// 使用 embed 加载模板和静态资源
	r.SetHTMLTemplate(template.Must(template.New("").Funcs(template.FuncMap{
		"formatDate": templatefuncs.FormatDate,
		"max":        templatefuncs.Max,
		"min":        templatefuncs.Min,
		"sub":        templatefuncs.Sub,
		"seq":        templatefuncs.Seq,
		"add":        templatefuncs.Add,
	}).ParseFS(templates, "templates/**/*.html"))) // 修改: 支持子目录中的模板文件
	subFS, _ := fs.Sub(assets, "assets")  //加载静态资源
	r.StaticFS("/assets", http.FS(subFS)) //修改: 支持子目录中的静态资源
	r.Static("/upload", models.ResolvePath("./upload")) //加载静态上传资源且不使用embed

	routers.IndexRoutersInit(r) // 初始前台路由
	routers.AdminRoutersInit(r) // 初始后台路由
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "index/404.html", gin.H{
			"title": "页面未找到",
		})
	})

	addr := fmt.Sprintf("%s:%s", config.App.Host, config.App.Port) // 构建服务地址

	r.Run(addr) // 启动服务
}
