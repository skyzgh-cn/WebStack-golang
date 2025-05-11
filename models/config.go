package models

import (
	"encoding/json"
	"io/ioutil"
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
	jsonConfig     *Config // 存储从config.json读取的配置
)

// 从config.json文件加载配置
func loadConfigFromFile(filePath string) (*Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// getEnv 获取环境变量，如果不存在则返回config.json中的值，如果config.json也不存在则返回默认值
func getEnv(key string, jsonValue string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if jsonValue != "" {
			return jsonValue
		}
		return defaultValue
	}
	return value
}

// getEnvBool 获取布尔类型的环境变量
func getEnvBool(key string, jsonValue bool, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return jsonValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return jsonValue
	}
	return boolValue
}

// LoadConfig 函数从环境变量和config.json加载配置
func LoadConfig() (*Config, error) {
	loadOnce.Do(func() {
		// 尝试从config.json加载配置
		fileConfig, err := loadConfigFromFile("config.json")
		if err != nil {
			// 如果加载失败，使用空配置
			fileConfig = &Config{}
		}
		jsonConfig = fileConfig

		config := &Config{}

		// 加载App配置
		config.App.Name = getEnv("APP_NAME", jsonConfig.App.Name, "WebStack")
		config.App.Host = getEnv("APP_HOST", jsonConfig.App.Host, "0.0.0.0")
		config.App.Port = getEnv("APP_PORT", jsonConfig.App.Port, "8080")
		config.App.Dbtype = getEnv("DB_TYPE", jsonConfig.App.Dbtype, "sqlite")

		// 加载MySQL配置
		config.Mysql.Type = "mysql"
		config.Mysql.Host = getEnv("MYSQL_HOST", jsonConfig.Mysql.Host, "127.0.0.1")
		config.Mysql.Port = getEnv("MYSQL_PORT", jsonConfig.Mysql.Port, "3306")
		config.Mysql.User = getEnv("MYSQL_USER", jsonConfig.Mysql.User, "webstack")
		config.Mysql.Password = getEnv("MYSQL_PASSWORD", jsonConfig.Mysql.Password, "webstack")
		config.Mysql.Database = getEnv("MYSQL_DATABASE", jsonConfig.Mysql.Database, "webstack")
		config.Mysql.Charset = getEnv("MYSQL_CHARSET", jsonConfig.Mysql.Charset, "utf8mb4")
		config.Mysql.ParseTime = getEnvBool("MYSQL_PARSE_TIME", jsonConfig.Mysql.ParseTime, true)
		config.Mysql.Loc = getEnv("MYSQL_LOC", jsonConfig.Mysql.Loc, "Local")

		// 加载SQLite配置
		config.Sqlite.Type = "sqlite"
		config.Sqlite.File = getEnv("SQLITE_FILE", jsonConfig.Sqlite.File, "./webstack.db")

		configInstance = config
	})

	return configInstance, nil
}
