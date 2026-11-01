package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
	"github.com/skyzgh-cn/WebStack-golang/models"
)

type SiteController struct{}

// maxLogoSize 限制站点 Logo 的最大字节数（2MB）。
const maxLogoSize = 2 << 20

// allowedLogoExt 允许上传的图片扩展名白名单。
var allowedLogoExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
}

// Settings 处理网站设置页面的 GET 请求
func (sc SiteController) Settings(c *gin.Context) {
	site := models.Site{}
	models.DB.First(&site)
	// 从数据库中读取网站设置

	c.HTML(200, "admin/settings.html", gin.H{
		"title":     "系统设置",
		"site":      site,
		"csrfToken": middleware.CSRFToken(c),
	})
}

// UpdateSettings 处理网站设置页面的 POST 请求
func (sc SiteController) UpdateSettings(c *gin.Context) {
	var site models.Site

	// 绑定表单数据（不包含文件）
	if err := c.ShouldBind(&site); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取当前数据库中的站点信息以获取旧的 Logo 路径
	var existingSite models.Site
	models.DB.First(&existingSite, site.Id)

	// 获取上传的文件
	file, err := c.FormFile("sitelogo")
	if err == nil {

		// 校验文件大小，避免超大文件耗尽磁盘
		if file.Size > maxLogoSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "图片大小不能超过 2MB"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if _, ok := allowedLogoExt[ext]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg、jpeg、png、gif、webp、bmp 格式的图片"})
			return
		}

		// 校验文件真实内容（magic number），防止仅改扩展名绕过
		if err := validateImageContent(file); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		uploadDir := models.ResolvePath("./upload")
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
				return
			}
		}

		// 使用随机文件名，避免上传的脚本被直接访问执行
		filename, err := randomLogoName(ext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成文件名失败"})
			return
		}
		dst := filepath.Join(uploadDir, filename)

		// 保存上传的文件
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败：" + err.Error()})
			return
		}

		// 清理此前上传的旧 Logo，避免目录堆积
		removeOldLogo(existingSite.Sitelogo)

		// 更新站点 Logo 路径
		site.Sitelogo = "/upload/" + filename
	} else {
		// 如果没有新上传文件，则保留旧的 Logo 路径
		site.Sitelogo = existingSite.Sitelogo
	}

	// 保存更新后的站点信息
	models.DB.Save(&site)

	// 重定向回设置页面
	c.Redirect(302, "/admin/settings")
}

// randomLogoName 生成随机且安全的 Logo 文件名。
func randomLogoName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("logo_%s%s", hex.EncodeToString(buf), ext), nil
}

// removeOldLogo 删除历史上传到 upload 目录的 Logo 文件。
func removeOldLogo(oldPath string) {
	if oldPath == "" || len(oldPath) < len("/upload/") || oldPath[:len("/upload/")] != "/upload/" {
		return
	}
	name := filepath.Base(oldPath)
	if name == "" || name == "." || name == ".." {
		return
	}
	os.Remove(filepath.Join(models.ResolvePath("./upload"), name))
}
