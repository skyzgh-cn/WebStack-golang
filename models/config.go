package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config 结构体用于存储应用程序配置
type Config struct {
	App struct {
		Name   string `json:"name"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		Dbtype string `json:"dbtype"` // 新增字段
		// SecureCookie 控制会话 Cookie 是否仅通过 HTTPS 传输。
		// 站点部署在 HTTPS 反向代理后时应置为 true。
		SecureCookie bool `json:"secureCookie"`
		// TrustedProxies 列出可信反向代理的 IP 或 CIDR。
		// 留空表示不信任任何代理，此时 X-Forwarded-* 头一律被忽略，
		// 可防止客户端伪造 X-Forwarded-Proto 绕过 HTTPS 判断。
		TrustedProxies []string `json:"trustedProxies"`
	} `json:"app"`
	Mysql struct {
		Type      string `json:"type"` // 可删，实际由dbtype决定
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
	loadErr        error
	loadOnce       sync.Once
)

// LoadConfig 函数用于从 config.json 文件加载配置
func LoadConfig() (*Config, error) {
	loadOnce.Do(func() { // 确保只加载一次配置
		path, err := findConfigFile()
		if err != nil {
			loadErr = err
			return
		}

		file, err := os.ReadFile(path) // 读取配置文件
		if err != nil {
			loadErr = fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
			return
		}

		// 记录配置文件目录，作为相对路径的基准
		configDir = filepath.Dir(path)

		var config Config
		if err := json.Unmarshal(file, &config); err != nil {
			loadErr = fmt.Errorf("解析配置文件失败: %w", err)
			return
		}

		// 环境变量优先于配置文件，避免敏感信息（如数据库密码）必须写入版本库
		applyEnvOverrides(&config)

		configInstance = &config // 缓存配置实例
	})

	if loadErr != nil {
		return nil, loadErr
	}
	return configInstance, nil
}

// findConfigFile 定位 config.json。
//
// 优先使用环境变量 WEBSTACK_CONFIG 指定的路径；否则依次尝试当前工作目录
// 及其各级父目录，以支持从任意目录启动程序。
func findConfigFile() (string, error) {
	if p := os.Getenv("WEBSTACK_CONFIG"); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("环境变量 WEBSTACK_CONFIG 指定的配置文件不存在: %s", p)
		}
		return p, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "config.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("未找到配置文件 config.json，请确认已上传该文件，或通过环境变量 WEBSTACK_CONFIG 指定路径")
}

// configDir 记录配置文件的所在目录，用于解析相对路径。
var configDir string

// BaseDir 返回程序运行的基础目录（配置文件所在目录）。
//
// 数据库文件、上传目录等相对路径均以此为基准，
// 保证从任意目录启动程序时行为一致。
func BaseDir() string {
	if configDir != "" {
		return configDir
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// resolvePath 将相对路径转换为基于配置目录的绝对路径。
func resolvePath(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(BaseDir(), path)
}

// ResolvePath 将相对路径转换为基于配置目录的绝对路径，供其他包使用。
func ResolvePath(path string) string {
	return resolvePath(path)
}

// applyEnvOverrides 使用环境变量覆盖配置项，便于在生产环境中安全地传递敏感信息。
func applyEnvOverrides(config *Config) {
	if v := os.Getenv("WEBSTACK_DB_TYPE"); v != "" {
		config.App.Dbtype = v
	}
	if v := os.Getenv("WEBSTACK_HOST"); v != "" {
		config.App.Host = v
	}
	if v := os.Getenv("WEBSTACK_PORT"); v != "" {
		config.App.Port = v
	}
	if v := os.Getenv("WEBSTACK_SECURE_COOKIE"); v != "" {
		config.App.SecureCookie = v == "true" || v == "1"
	}
	if v := os.Getenv("WEBSTACK_TRUSTED_PROXIES"); v != "" {
		config.App.TrustedProxies = strings.Split(v, ",")
	}

	if v := os.Getenv("WEBSTACK_MYSQL_HOST"); v != "" {
		config.Mysql.Host = v
	}
	if v := os.Getenv("WEBSTACK_MYSQL_PORT"); v != "" {
		config.Mysql.Port = v
	}
	if v := os.Getenv("WEBSTACK_MYSQL_USER"); v != "" {
		config.Mysql.User = v
	}
	if v := os.Getenv("WEBSTACK_MYSQL_PASSWORD"); v != "" {
		config.Mysql.Password = v
	}
	if v := os.Getenv("WEBSTACK_MYSQL_DATABASE"); v != "" {
		config.Mysql.Database = v
	}

	if v := os.Getenv("WEBSTACK_SQLITE_FILE"); v != "" {
		config.Sqlite.File = v
	}
}
