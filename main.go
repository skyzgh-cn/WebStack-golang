package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/routers"
	"html/template"
	"io/fs"
	"net/http"
	"os"
)

//go:embed templates/**/*
var templates embed.FS

//go:embed assets/*
var assets embed.FS

// Config 结构体用于存储应用程序配置
type Config struct {
	App struct {
		Name string `json:"name"` // 应用名称
		Host string `json:"host"` // 应用主机地址
		Port string `json:"port"` // 应用端口号
	} `json:"app"`
}

// loadConfig 函数用于从 config.json 文件加载配置
func loadConfig() (*Config, error) {
	file, err := os.ReadFile("config.json") // 读取配置文件
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config) // 解析 JSON 数据到 Config 结构体
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	config, err := loadConfig() // 加载配置
	if err != nil {
		fmt.Println("无法加载配置文件:", err) // 如果加载失败，打印错误信息
		return
	}

	r := gin.Default() // 创建 Gin 引擎实例
	// 使用 embed 加载模板和静态资源
	r.SetHTMLTemplate(template.Must(template.ParseFS(templates, "templates/**/*.html"))) // 修改: 支持子目录中的模板文件

	subFS, _ := fs.Sub(assets, "assets")
	r.StaticFS("/assets", http.FS(subFS))

	routers.IndexRoutersInit(r) // 初始前台路由

	addr := fmt.Sprintf("%s:%s", config.App.Host, config.App.Port) // 构建服务地址
	fmt.Printf("服务即将运行在: http://%s\n", addr)                       // 打印服务运行地址

	r.Run(addr) // 启动服务
}
