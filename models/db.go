package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strings" // Add this import statement
	"time"
)

var DB *gorm.DB

func init() {
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("无法加载配置文件: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config.Db.User,
		config.Db.Password,
		config.Db.Host,
		config.Db.Port,
		config.Db.Database,
		config.Db.Charset,
		config.Db.ParseTime,
		config.Db.Loc,
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取底层 SQL DB 失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(50)                  // 最大连接数
	sqlDB.SetMaxIdleConns(20)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期

	fmt.Println("数据库连接成功")

	// 自动迁移数据库表
	err = DB.AutoMigrate(&Admin{}, &Website{}, &Group{}, &Site{})
	if err != nil {
		log.Println("数据库迁移失败:", err)

	}

	log.Println("数据库迁移成功")

	var adminCount int64
	var siteCount int64
	DB.Model(&Admin{}).Count(&adminCount)
	DB.Model(&Site{}).Count(&siteCount)
	if adminCount == 0 || siteCount == 0 {

		site := Site{
			Sitename:    "SkyZgh网址导航",
			Siteurl:     "http://www.skyzgh.com",
			Sitelogo:    "/assets/images/logo@2x.png",
			Description: "WebStack 是一个基于Go语言开发的开源项目，用于快速搭建个人导航站",
			Keywords:    "skyzgh,skyzgh网址导航,skyzgh导航,skyzgh网址导航系统",
			Aboutweb:    "<blockquote><p>❤️基于Golang开源的网址导航网站项目，具备完整的前后台，您可以拿来制作自己平日收藏的网址导航。</p><p>如果你也是开发者，如果你也正好喜欢折腾，那希望这个网站能给你带来一些作用。</p><p>前端使用WebStackPage：<a href=\"https://github.com/WebStackPage/WebStackPage.github.io\">前端开源地址</a></p><p>本站开源地址：<a href=\"https://github.com/skyzgh-cn/WebStack-golang\">WebStack-golang</a></p></blockquote>",
			Aboutme:     "                                <div class=\"col-sm-4\">\n                                    <div class=\"xe-widget xe-conversations box2 label-info\" onclick=\"window.open('http://www.skyzgh.com/', '_blank')\" data-toggle=\"tooltip\" data-placement=\"bottom\" title=\"\" data-original-title=\"http://www.skyzgh.com/\">\n                                        <div class=\"xe-comment-entry\">\n                                            <a class=\"xe-user-img\">\n                                                <img src=\"../assets/images/favicon.png\" class=\"img-circle\" width=\"40\">\n                                            </a>\n                                            <div class=\"xe-comment\">\n                                                <a href=\"#\" class=\"xe-user-name overflowClip_1\">\n                                                    <strong>SkyZgh</strong>\n                                                </a>\n                                                <p class=\"overflowClip_2\">www.skyzgh.com</p>\n                                            </div>\n                                        </div>\n                                    </div>\n                                </div>\n                                <div class=\"col-md-8\">\n                                    <div class=\"row\">\n                                        <div class=\"col-sm-12\">\n                                            <br />\n                                            <blockquote>\n                                                <p>\n                                                    这是一个公益项目，而且是<a href=\"https://github.com/skyzgh-cn/WebStack-golang\"> 开源 </a>的。你也可以拿来制作自己的网址导航。如果你有更好的想法，可以通过个人网站<a href=\"http://www.skyzgh.com\"><span class=\"label label-info\" data-toggle=\"tooltip\" data-placement=\"left\" title=\"\" data-original-title=\"Hello I am SkyZgh\">skyzgh.com</span></a>中的联系方式找到我，欢迎与我交流分享。\n                                                </p>\n                                            </blockquote>\n                                        </div>\n                                    </div>\n                                    <br>\n                                </div>",
			Copyright:   "<div class=\"footer-text\">&copy; 2025 - 2030<a href=\"about\"><strong>WebStack-golang</strong></a> design by <a href=\"http://www.skyzgh.com\" target=\"_blank\"><strong>SkyZgh</strong></a></div>",
			Count:       "<script>var _hmt = _hmt || [];(function() {  var hm = document.createElement(\"script\");  hm.src = \"https://hm.baidu.com/hm.js?19dc88cd2eab7c6e00f684e51aebce05\";  var s = document.getElementsByTagName(\"script\")[0];   s.parentNode.insertBefore(hm, s);})();</script>",
		}
		if err := DB.Create(&site).Error; err != nil {
			log.Println("创建默认站点信息失败:", err)
		} else {
			log.Println("已创建默认站点信息")
		}
		//批量执行default.sql文件，新增默认数据到各个表
		file, err := os.ReadFile("./default.sql")
		if err != nil {
			log.Println("读取default.sql文件失败:", err)
			return
		}
		// 按分号分割 SQL 语句
		statements := strings.Split(string(file), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				log.Printf("即将执行的 SQL 语句: %s", stmt)
				if err := DB.Exec(stmt).Error; err != nil {
					log.Printf("执行 SQL 语句失败: %s, 错误信息: %v", stmt, err)
					return
				}
			}
		}
		log.Println("default.sql文件执行成功")

	}

	//// 创建默认管理员（如果没有任何管理员存在）
	//var adminCount int64
	//DB.Model(&Admin{}).Count(&adminCount)
	//if adminCount == 0 {
	//	defaultAdmin := Admin{
	//		Username: "admin",
	//		Password: "21232f297a57a5a743894a0e4a801fc3", // md5("admin")
	//	}
	//	if err := DB.Create(&defaultAdmin).Error; err != nil {
	//		log.Println("创建默认管理员失败: %v", err)
	//	} else {
	//		log.Println("已创建默认管理员账号")
	//	}
	//} else {
	//	log.Println("管理员账号已存在，跳过创建默认管理员")
	//}
	//
	////检查是否存在site信息
	//var count2 int64
	//DB.Model(&Site{}).Count(&count2)
	//if count2 == 0 {
	//	site := Site{
	//		Sitename:    "SkyZgh网址导航",
	//		Siteurl:     "http://www.skyzgh.com",
	//		Sitelogo:    "/assets/images/logo@2x.png",
	//		Description: "WebStack 是一个基于Go语言开发的开源项目，用于快速搭建个人导航站",
	//		Keywords:    "skyzgh,skyzgh网址导航,skyzgh导航,skyzgh网址导航系统",
	//		Aboutweb:    "<blockquote><p>❤️基于Golang开源的网址导航网站项目，具备完整的前后台，您可以拿来制作自己平日收藏的网址导航。</p><p>如果你也是开发者，如果你也正好喜欢折腾，那希望这个网站能给你带来一些作用。</p></blockquote>",
	//		Aboutme:     "<divclass=\"col-sm-4\"><divclass=\"xe-widgetxe-conversationsbox2label-info\"onclick=\"window.open('http://www.skyzgh.com/','_blank')\"data-toggle=\"tooltip\"data-placement=\"bottom\"title=\"\"data-original-title=\"http://www.skyzgh.com/\"><divclass=\"xe-comment-entry\"><aclass=\"xe-user-img\"><imgsrc=\"../assets/images/favicon.png\"class=\"img-circle\"width=\"40\"></a><divclass=\"xe-comment\"><ahref=\"#\"class=\"xe-user-nameoverflowClip_1\"><strong>SkyZgh</strong></a><pclass=\"overflowClip_2\">www.skyzgh.com</p></div></div></div></div><divclass=\"col-md-8\"><divclass=\"row\"><divclass=\"col-sm-12\"><br/><blockquote><p>这是一个公益项目，而且是<ahref=\"https://github.com/skyzgh-cn/WebStack-golang\">开源</a>的。你也可以拿来制作自己的网址导航。如果你有更好的想法，可以通过个人网站<ahref=\"http://www.skyzgh.com\"><spanclass=\"labellabel-info\"data-toggle=\"tooltip\"data-placement=\"left\"title=\"\"data-original-title=\"HelloIamaTooltip\">skyzgh.com</span></a>中的联系方式找到我，欢迎与我交流分享。</p></blockquote></div></div><br></div>",
	//		Copyright:   "<div class=\"footer-text\">&copy; 2025 - 2030<a href=\"about\"><strong>WebStack-golang</strong></a> design by <a href=\"http://www.skyzgh.com\" target=\"_blank\"><strong>SkyZgh</strong></a></div>",
	//		Count:       "<script>var _hmt = _hmt || [];(function() {  var hm = document.createElement(\"script\");  hm.src = \"https://hm.baidu.com/hm.js?19dc88cd2eab7c6e00f684e51aebce05\";  var s = document.getElementsByTagName(\"script\")[0];   s.parentNode.insertBefore(hm, s);})();</script>",
	//	}
	//	if err := DB.Create(&site).Error; err != nil {
	//		log.Println("创建默认站点信息失败:", err)
	//	} else {
	//		log.Println("已创建默认站点信息")
	//	}
	//}
}
