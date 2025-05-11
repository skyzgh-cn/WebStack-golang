package models

import (
	"os"
	"strconv"
	"sync"
)

// Config 结构体用于存储应用程序配置
type Config struct {
	App struct {
		Name   string `json:"name"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		Dbtype string `json:"dbtype"`
	} `json:"app"`
	Mysql struct {
		Type      string `json:"type"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		User      string `json:"user"`
		Password  string `json:"password"`
		Database  string `json:"database"`
		Charset   string `json:"charset"`
		ParseTime bool   `json:"parseTime"`
		Loc       string `json:"loc"`
	} `json:"mysql"`
	Sqlite struct {
		Type string `json:"type"`
		File string `json:"file"`
	} `json:"sqlite"`
}

// 全局变量，用于存储单例实例
var (
	configInstance *Config
	loadOnce       sync.Once
)

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvBool 获取布尔类型的环境变量
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// LoadConfig 函数从环境变量加载配置
func LoadConfig() (*Config, error) {
	loadOnce.Do(func() {
		config := &Config{}

		// 加载App配置
		config.App.Name = getEnv("APP_NAME", "WebStack")
		config.App.Host = getEnv("APP_HOST", "0.0.0.0")
		config.App.Port = getEnv("APP_PORT", "8080")
		config.App.Dbtype = getEnv("DB_TYPE", "sqlite")

		// 加载MySQL配置
		config.Mysql.Type = "mysql"
		config.Mysql.Host = getEnv("MYSQL_HOST", "127.0.0.1")
		config.Mysql.Port = getEnv("MYSQL_PORT", "3306")
		config.Mysql.User = getEnv("MYSQL_USER", "webstack")
		config.Mysql.Password = getEnv("MYSQL_PASSWORD", "webstack")
		config.Mysql.Database = getEnv("MYSQL_DATABASE", "webstack")
		config.Mysql.Charset = getEnv("MYSQL_CHARSET", "utf8mb4")
		config.Mysql.ParseTime = getEnvBool("MYSQL_PARSE_TIME", true)
		config.Mysql.Loc = getEnv("MYSQL_LOC", "Local")

		// 加载SQLite配置
		config.Sqlite.Type = "sqlite"
		config.Sqlite.File = getEnv("SQLITE_FILE", "./webstack.db")

		configInstance = config
	})

	return configInstance, nil
}
