package models

import (
	"encoding/json"
	"os"
	"sync"
)

// Config 结构体用于存储应用程序配置
type Config struct {
	App struct {
		Name string `json:"name"` // 应用名称
		Host string `json:"host"` // 应用主机地址
		Port string `json:"port"` // 应用端口号
	} `json:"app"`
	Db struct { // 数据库配置
		Type      string `json:"type"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		User      string `json:"user"`
		Password  string `json:"password"`
		Database  string `json:"database"`
		Charset   string `json:"charset"`
		ParseTime bool   `json:"parseTime"`
		Loc       string `json:"loc"`
	} `json:"db"`
}

// 全局变量，用于存储单例实例
var (
	configInstance *Config
	loadOnce       sync.Once
)

// LoadConfig 函数用于从 config.json 文件加载配置
func LoadConfig() (*Config, error) {
	loadOnce.Do(func() { // 确保只加载一次配置
		file, err := os.ReadFile("./config.json") // 读取配置文件
		if err != nil {
			return
		}

		var config Config
		err = json.Unmarshal(file, &config) // 解析 JSON 数据到 Config 结构体
		if err != nil {
			return
		}

		configInstance = &config // 缓存配置实例
	})

	return configInstance, nil // 返回缓存的实例
}
