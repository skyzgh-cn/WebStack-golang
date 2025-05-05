package admin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/models"
	"golang.org/x/net/html/charset"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type WebsiteController struct {
}

func (wc WebsiteController) Index(c *gin.Context) {
	var websites []models.Website
	models.DB.Preload("Group").Find(&websites)
	var groups []models.Group
	models.DB.Find(&groups)
	c.HTML(http.StatusOK, "admin/websites.html", gin.H{
		"title":    "网站管理",
		"websites": websites,
		"groups":   groups,
	})

	//c.JSON(200, gin.H{
	//	"websites": websites,
	//})
}

func (wc WebsiteController) Save(c *gin.Context) {

	ID := c.PostForm("id")
	var website models.Website
	if ID == "0" {

		if err := c.ShouldBind(&website); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if website.Logo == "" {
			website.Logo = "/assets/images/default.png"
		}
		models.DB.Create(&website)

	} else {

		if err := c.ShouldBind(&website); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if website.Logo == "" {
			website.Logo = "/assets/images/default.png"
		}
		website.CreatedAt = time.Now()
		models.DB.Save(&website)

	}

	c.Redirect(302, "/admin/websites")
}

func (wc WebsiteController) Delete(c *gin.Context) {
	id := c.PostForm("id")
	models.DB.Delete(&models.Website{}, id)
	c.JSON(200, gin.H{"code": 0, "msg": "删除成功"})
}

// 抓取接口
type FetchResponse struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg,omitempty"`
	Title       string `json:"title,omitempty"`
	Favicon     string `json:"favicon,omitempty"` // 新增
	Description string `json:"description,omitempty"`
}

func (wc WebsiteController) FetchMeta(c *gin.Context) {
	type Request struct {
		URL string `json:"url" binding:"required"`
	}
	var req Request

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, FetchResponse{Code: -1, Msg: "参数错误：" + err.Error()})
		return
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, FetchResponse{Code: -1, Msg: "无效的 URL 格式"})
		return
	}

	ip := net.ParseIP(parsedURL.Hostname())
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		c.JSON(http.StatusBadRequest, FetchResponse{Code: -1, Msg: "不允许访问局域网或本机地址"})
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, FetchResponse{Code: -1, Msg: "请求目标网站失败：" + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, FetchResponse{Code: -1, Msg: "目标网站返回状态码：" + resp.Status})
		return
	}

	body, _ := io.ReadAll(resp.Body)
	reader := bytes.NewReader(body)
	htmlContent, err := charset.NewReaderLabel("", reader)
	if err != nil {
		htmlContent = strings.NewReader(string(body))
	}
	content, _ := io.ReadAll(htmlContent)
	html := string(content)

	// 提取 title
	title := extractRegex(html, `<title\b[^>]*>(.*?)</title>`, true)
	if title == "" {
		title = extractRegex(html, `property="og:title"\s+content="(.*?)"`, false)
	}

	// 提取 description
	desc := extractRegex(html, `<meta[^>]+name\s*=\s*["']description["'][^>]*content\s*=\s*["'](.*?)["']`, false)
	if desc == "" {
		desc = extractRegex(html, `property="og:description"\s+content="(.*?)"`, false)
	}

	// 提取 favicon
	const faviconPattern = `(?i)<link\b[^>]*?\b(?:rel\s*=\s*["'](?:shortcut\s+)?icon["'])[^>]*?href\s*=\s*["'](.*?)["'][^>]*?>`
	favicon := extractRegex(html, faviconPattern, false)

	var faviconURL string // ✅ 在外层声明，避免 undefined 报错

	if favicon != "" {
		if strings.HasPrefix(favicon, "//") {
			favicon = "https:" + favicon
		}

		// 判断是否是完整 URL
		parsedFavicon, err := url.Parse(favicon)
		if err == nil && (parsedFavicon.Scheme == "http" || parsedFavicon.Scheme == "https") {
			// 如果是完整 URL，直接使用
			faviconURL = favicon
		} else {
			// 否则拼接基础 URL
			base, _ := url.Parse(req.URL)
			faviconURL = base.ResolveReference(&url.URL{Path: favicon}).String()
		}
	} else {
		// ✅ 在 else 分支中也赋值
		faviconURL = "/assets/images/default.png"
	}

	// ✅ 最终返回 JSON
	c.JSON(http.StatusOK, FetchResponse{
		Code:        0,
		Title:       title,
		Favicon:     faviconURL,
		Description: desc,
	})
}

// 提取 HTML 内容通用函数
func extractRegex(html, pattern string, stripTags bool) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		result := strings.TrimSpace(match[1])
		if stripTags {
			return regexp.MustCompile(`<[^>]+>`).ReplaceAllString(result, "")
		}
		return result
	}
	return ""
}
