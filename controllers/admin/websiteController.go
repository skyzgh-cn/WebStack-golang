package admin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skyzgh-cn/WebStack-golang/middleware"
	"github.com/skyzgh-cn/WebStack-golang/models"
	"golang.org/x/net/html/charset"
)

type WebsiteController struct {
}

func (wc WebsiteController) Index(c *gin.Context) {
	// 获取分组ID参数
	groupID := c.Query("group_id")
	// 获取名称搜索参数
	searchName := c.Query("name")

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20 // 默认每页20条

	// 校验 page 是否为正数（须在计算 offset 之前校正，避免负数偏移）
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 查询网站数据（带分组筛选和名称搜索）
	var websites []models.Website
	query := models.DB.Preload("Group")
	if groupID != "" && groupID != "0" {
		query = query.Where("group_id = ?", groupID)
	}
	if searchName != "" {
		query = query.Where("name LIKE ?", "%"+searchName+"%")
	}
	query = query.Order("created_at DESC")
	query.Limit(pageSize).Offset(offset).Find(&websites)

	// 查询总记录数（不要带Limit和Offset）
	var total int64
	countQuery := models.DB.Model(&models.Website{})
	if groupID != "" && groupID != "0" {
		countQuery = countQuery.Where("group_id = ?", groupID)
	}
	if searchName != "" {
		countQuery = countQuery.Where("name LIKE ?", "%"+searchName+"%")
	}
	countQuery.Count(&total)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// 校验 totalPages 是否为正数
	if totalPages <= 0 {
		totalPages = 1
	}
	// 添加调试信息
	fmt.Printf("分页信息:\n")
	fmt.Printf("- 当前页码: %d\n", page)
	fmt.Printf("- 总页数: %d\n", totalPages)
	// 查询所有分组（用于下拉菜单）
	var groups []models.Group
	models.DB.Find(&groups)

	// 渲染模板
	c.HTML(http.StatusOK, "admin/websites.html", gin.H{
		"title":            "网站管理",
		"websites":         websites,
		"groups":           groups,
		"current_group_id": groupID,
		"search_name":      searchName,
		"page":             page,
		"total_pages":      totalPages,
		"csrfToken":        middleware.CSRFToken(c),
	})
}

func (wc WebsiteController) Save(c *gin.Context) {

	ID := c.PostForm("id")
	var website models.Website
	if err := c.ShouldBind(&website); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if website.Logo == "" {
		website.Logo = "/assets/images/default.png"
	}

	if ID == "0" {
		// 新增网站：由数据库生成自增主键与创建时间
		website.Id = 0
		models.DB.Create(&website)
	} else {
		// 更新网站：保留原有创建时间，避免编辑操作刷新"添加时间"
		var existingWebsite models.Website
		if err := models.DB.First(&existingWebsite, ID).Error; err != nil {
			c.JSON(404, gin.H{"error": "网站不存在"})
			return
		}
		existingWebsite.Name = website.Name
		existingWebsite.GroupId = website.GroupId
		existingWebsite.Url = website.Url
		existingWebsite.Logo = website.Logo
		existingWebsite.Description = website.Description
		models.DB.Save(&existingWebsite)
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

	parsedURL, err := validateTargetURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, FetchResponse{Code: -1, Msg: err.Error()})
		return
	}

	// 预先解析域名并确认不指向内网
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := resolveAndCheckHost(ctx, parsedURL.Hostname()); err != nil {
		c.JSON(http.StatusBadRequest, FetchResponse{Code: -1, Msg: err.Error()})
		return
	}

	client := newSafeHTTPClient()

	resp, err := client.Get(parsedURL.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, FetchResponse{Code: -1, Msg: "请求目标网站失败：" + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, FetchResponse{Code: -1, Msg: "目标网站返回状态码：" + resp.Status})
		return
	}

	// 限制响应体大小，防止内存耗尽
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxHTMLBodySize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, FetchResponse{Code: -1, Msg: "读取目标网站内容失败"})
		return
	}
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

	// 新增：尝试直接抓取 favicon.ico 或 favicon.png
	var faviconURL string
	faviconPaths := []string{"/favicon.ico", "/favicon.png"} // 默认的 favicon 路径
	for _, path := range faviconPaths {
		candidate := parsedURL.Scheme + "://" + parsedURL.Host + path // 拼接到完整 URL
		favResp, err := client.Get(candidate)                         // 尝试抓取
		if err == nil {
			// 即使成功也要关闭响应体，避免连接泄漏
			io.Copy(io.Discard, io.LimitReader(favResp.Body, 1024))
			favResp.Body.Close()
			if favResp.StatusCode == 200 {
				faviconURL = candidate
				break // 使用当前 faviconURL
			}
		}
		faviconURL = "" // 如果抓取失败，清空 faviconURL
	}

	// 如果直接抓取失败，进入现有的 favicon 提取逻辑
	if faviconURL == "" {
		const faviconPattern = `(?i)<link\b[^>]*?\b(?:rel\s*=\s*["'](?:shortcut\s+)?icon["'])[^>]*?href\s*=\s*["'](.*?)["'][^>]*?>`
		favicon := extractRegex(html, faviconPattern, false)

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
			// 如果仍然没有找到 favicon，使用默认的 favicon
			faviconURL = "/assets/images/default.png"
		}
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
