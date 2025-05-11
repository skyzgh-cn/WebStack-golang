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

func init() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置文件: %v", err)
	}

	switch config.App.Dbtype {
	case "mysql":
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
		dir := filepath.Dir(config.Sqlite.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Println("无法创建目录 %s: %v", dir, err)
		}
		DB, err = gorm.Open(sqlite.Open(config.Sqlite.File+"?cache=shared&_sync=OFF"), &gorm.Config{
			SkipDefaultTransaction: true,
			DisableAutomaticPing:   true,
		})

		if err != nil {
			log.Fatalf("数据库连接失败: %v", err)
		}
	default:
		log.Fatalf("不支持的数据库类型: %s", config.App.Dbtype)
	}

	// 仅对 MySQL 设置连接池参数
	if config.App.Dbtype == "mysql" {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Fatalf("获取底层 SQL DB 失败: %v", err)
		}
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
	}

	fmt.Println("数据库连接成功")

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&Admin{}, &Website{}, &Group{}, &Site{})
	if err != nil {
		log.Println("数据库迁移失败:", err)
	} else {
		log.Println("数据库迁移成功")
	}

	// 检查并初始化默认数据
	initializeDefaultData()
}

// 初始化默认数据（站点信息 + SQL 文件执行）
func initializeDefaultData() {
	var adminCount, siteCount int64
	DB.Model(&Admin{}).Count(&adminCount)
	DB.Model(&Site{}).Count(&siteCount)

	if adminCount == 0 || siteCount == 0 {
		// 创建默认站点信息
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

		// 执行 default.sql 文件
		file, err := os.ReadFile("./default.sql")
		if err != nil {
			log.Println("读取 default.sql 文件失败:", err)
			return
		}

		statements := strings.Split(string(file), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				log.Printf("即将执行 SQL 语句: %s", stmt)
				if err := DB.Exec(stmt).Error; err != nil {
					log.Printf("执行 SQL 语句失败: %s, 错误信息: %v", stmt, err)
					return
				}
			}
		}
		log.Println("default.sql 文件执行成功")
	}
}
