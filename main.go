package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
	"github.com/skyzgh-cn/WebStack-golang/routers"
	"github.com/skyzgh-cn/WebStack-golang/templatefuncs"
)

//go:embed templates/**/*
var templates embed.FS

//go:embed assets/*
var assets embed.FS

func main() {
	config, err := models.LoadConfig() // 加载配置文件
	if err != nil {
		fmt.Println("无法加载配置文件:", err) // 如果加载失败，打印错误信息
		return
	}

	r := gin.Default() // 创建 Gin 引擎实例

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
	r.Static("/upload", "./upload")       //加载静态上传资源且不使用embed

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
