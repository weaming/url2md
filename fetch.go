package url2md

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	readability "codeberg.org/readeck/go-readability/v2"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

const DEFAULT_TIMEOUT = 30 * time.Second

// convertHTML 调用第三方库完成 HTML 到 Markdown 的转换
func convertHTML(rawHTML string) (string, error) {
	return htmltomarkdown.ConvertString(rawHTML)
}

// Fetch 抓取网页并提取正文与元信息
func Fetch(rawURL string) (*Page, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("无效的 URL: %s", rawURL)
	}

	client := req.C().
		ImpersonateChrome().
		EnableInsecureSkipVerify().
		SetTimeout(DEFAULT_TIMEOUT)

	resp, err := client.R().Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("请求网页失败: %w", err)
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("网页返回异常状态码: %d", resp.StatusCode)
	}

	return Parse(resp.String(), rawURL)
}

// FetchMarkdown 抓取网页并直接返回正文 Markdown
func FetchMarkdown(rawURL string) (string, error) {
	page, err := Fetch(rawURL)
	if err != nil {
		return "", err
	}
	return page.Markdown()
}

// Parse 从原始 HTML 中提取元信息与正文，pageURL 用于解析相对链接
func Parse(rawHTML, pageURL string) (*Page, error) {
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("解析 URL 失败: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	page := &Page{URL: pageURL}
	page.fillMeta(doc, baseURL)

	// 站点规则命中时优先使用，失败再退回通用提取
	if rule := findRule(baseURL); rule != nil {
		if rawHTML, text := rule.extract(doc); rawHTML != "" {
			page.HTML = rawHTML
			page.Text = text

			if title := firstText(doc, rule.Title); title != "" {
				page.Title = title
			}
			if author := firstText(doc, rule.Author); author != "" {
				page.Author = author
			}

			return page, nil
		}
	}

	article, err := readability.FromReader(strings.NewReader(rawHTML), baseURL)
	if err != nil {
		return nil, fmt.Errorf("提取正文失败: %w", err)
	}

	var textBuf strings.Builder
	if renderErr := article.RenderText(&textBuf); renderErr == nil {
		page.Text = textBuf.String()
	}

	var htmlBuf strings.Builder
	if renderErr := article.RenderHTML(&htmlBuf); renderErr == nil {
		page.HTML = htmlBuf.String()
	}

	// 正文提取失败时退化为标题，保证调用方至少有内容可展示
	if strings.TrimSpace(page.Text) == "" {
		page.Text = page.Title
		page.HTML = page.Title
	}

	return page, nil
}
