package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// isMySQL 标记当前使用的数据库类型，用于适配 SQL 方言差异。
var isMySQL bool

// InitDB 建立数据库连接、迁移表结构并初始化默认数据。
//
// 由 main 显式调用，而不是放在 init() 中，避免包被导入时即产生副作用
// （例如读取配置、创建数据库文件），同时也便于测试注入临时数据库。
func InitDB() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("无法加载配置文件: %w", err)
	}

	switch config.App.Dbtype {
	case "mysql":
		isMySQL = true
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
			config.Mysql.User,
			config.Mysql.Password,
			config.Mysql.Host,
			config.Mysql.Port,
			config.Mysql.Database,
			config.Mysql.Charset,
			config.Mysql.ParseTime,
			config.Mysql.Loc,
		)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
		})
	case "sqlite":
		// 确保目录存在
		sqlitePath := resolvePath(config.Sqlite.File)
		dir := filepath.Dir(sqlitePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("无法创建目录 %s: %v", dir, err)
		}
		DB, err = gorm.Open(sqlite.Open(sqlitePath+"?cache=shared&_sync=OFF"), &gorm.Config{
			SkipDefaultTransaction: true,
			DisableAutomaticPing:   true,
		})
	default:
		return fmt.Errorf("不支持的数据库类型: %s", config.App.Dbtype)
	}

	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 仅对 MySQL 设置连接池参数
	if config.App.Dbtype == "mysql" {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("获取底层 SQL DB 失败: %w", err)
		}
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	fmt.Println("数据库连接成功")

	// 自动迁移数据库表结构
	if err := DB.AutoMigrate(&Admin{}, &Website{}, &Group{}, &Site{}); err != nil {
		log.Println("数据库迁移失败:", err)
	} else {
		log.Println("数据库迁移成功")
	}

	// 检查并初始化默认数据
	initializeDefaultData()

	return nil
}

// 初始化默认数据（站点信息 + 默认管理员 + SQL 文件执行）
func initializeDefaultData() {
	var adminCount, siteCount, groupCount int64
	DB.Model(&Admin{}).Count(&adminCount)
	DB.Model(&Site{}).Count(&siteCount)
	DB.Model(&Group{}).Count(&groupCount)

	// 站点信息：仅在缺失时创建，避免重复插入
	if siteCount == 0 {
		site := Site{
			Sitename:    "SkyZgh网址导航",
			Siteurl:     "http://www.skyzgh.com",
			Sitelogo:    "/assets/images/logo@2x.png",
			Description: "WebStack 是一个基于Go语言开发的开源项目，用于快速搭建个人导航站",
			Keywords:    "skyzgh,skyzgh网址导航,skyzgh导航,skyzgh网址导航系统",
			Aboutweb: `<blockquote><p>❤️基于Golang开源的网址导航网站项目，具备完整的前后台，您可以拿来制作自己平日收藏的网址导航。</p>
<p>如果你也是开发者，如果你也正好喜欢折腾，那希望这个网站能给你带来一些作用。</p>
<p>前端使用WebStackPage：<a href="https://github.com/WebStackPage/WebStackPage.github.io">前端开源地址</a></p>
<p>本站开源地址：<a href="https://github.com/skyzgh-cn/WebStack-golang">WebStack-golang</a></p></blockquote>`,
			Aboutme: `<div class="col-sm-4">
  <div class="xe-widget xe-conversations box2 label-info" onclick="window.open('http://www.skyzgh.com/', '_blank')" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="http://www.skyzgh.com/">
    <div class="xe-comment-entry">
      <a class="xe-user-img">
        <img src="../assets/images/favicon.png" class="img-circle" width="40">
      </a>
      <div class="xe-comment">
        <a href="#" class="xe-user-name overflowClip_1">
          <strong>SkyZgh</strong>
        </a>
        <p class="overflowClip_2">www.skyzgh.com</p>
      </div>
    </div>
  </div>
</div>
<div class="col-md-8">
  <div class="row">
    <div class="col-sm-12">
      <br />
      <blockquote>
        <p>
          这是一个公益项目，而且是<a href="https://github.com/skyzgh-cn/WebStack-golang"> 开源 </a>的。你也可以拿来制作自己的网址导航。如果你有更好的想法，可以通过个人网站
          <a href="http://www.skyzgh.com">
            <span class="label label-info" data-toggle="tooltip" data-placement="left" title="" data-original-title="Hello I am SkyZgh">skyzgh.com</span>
          </a>中的联系方式找到我，欢迎与我交流分享。
        </p>
      </blockquote>
    </div>
  </div>
  <br>
</div>`,
			Copyright: `<div class="footer-text">&copy; 2025 - 2030<a href="about"><strong>WebStack-golang</strong></a> design by <a href="http://www.skyzgh.com" target="_blank"><strong>SkyZgh</strong></a></div>`,
			Count: `<script>var _hmt = _hmt || [];(function() {  
  var hm = document.createElement("script");  
  hm.src = "https://hm.baidu.com/hm.js?19dc88cd2eab7c6e00f684e51aebce05";  
  var s = document.getElementsByTagName("script")[0];   
  s.parentNode.insertBefore(hm, s);
})();</script>`,
		}

		if err := DB.Create(&site).Error; err != nil {
			log.Println("创建默认站点信息失败:", err)
		} else {
			log.Println("已创建默认站点信息")
		}
	}

	// 默认管理员：使用 bcrypt 存储密码
	if adminCount == 0 {
		hashed, err := HashPassword("admin")
		if err != nil {
			log.Println("生成默认管理员密码失败:", err)
		} else if err := DB.Create(&Admin{Username: "admin", Password: hashed}).Error; err != nil {
			log.Println("创建默认管理员失败:", err)
		} else {
			log.Println("已创建默认管理员 admin/admin")
		}
	}

	// 分类与网站数据：仅在分类表为空时执行 default.sql
	if groupCount > 0 {
		return
	}

	file, err := os.ReadFile(resolvePath("default.sql"))
	if err != nil {
		log.Println("读取 default.sql 文件失败:", err)
		return
	}

	statements := strings.Split(string(file), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// 使用 INSERT IGNORE 语义，保证重复执行时不会因主键冲突中断
		insertStmt := makeInsertIgnore(stmt)
		if err := DB.Exec(insertStmt).Error; err != nil {
			log.Printf("执行 SQL 语句失败, 错误信息: %v", err)
			return
		}
	}
	log.Println("default.sql 文件执行成功")
}

// makeInsertIgnore 将 INSERT 语句改写为忽略重复主键的形式。
//
// default.sql 使用固定主键插入种子数据，若因并发启动等原因被重复执行，
// 原始语句会因主键冲突报错并中断后续语句。这里按数据库方言改写为
// INSERT IGNORE / INSERT OR IGNORE，使其具备幂等性。
func makeInsertIgnore(stmt string) string {
	upper := strings.ToUpper(stmt)
	if !strings.HasPrefix(upper, "INSERT INTO ") {
		return stmt
	}

	if strings.Contains(upper, " OR IGNORE") || strings.Contains(upper, " IGNORE ") {
		return stmt
	}

	if isMySQL {
		return strings.Replace(stmt, "INSERT INTO ", "INSERT IGNORE INTO ", 1)
	}
	return strings.Replace(stmt, "INSERT INTO ", "INSERT OR IGNORE INTO ", 1)
}
