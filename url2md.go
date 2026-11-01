package url2md

import (
	"fmt"
	"strings"
)

// Page 表示一个抓取并解析完成的网页
type Page struct {
	URL string

	Title       string
	Author      string
	Description string
	SiteName    string
	Keywords    []string
	ImageURL    string
	ImageWidth  int
	ImageHeight int
	IconURL     string

	// Text 由 readability 提取的正文纯文本
	Text string
	// HTML 由 readability 提取的正文 HTML
	HTML string
}

// Markdown 把正文 HTML 转换为 Markdown
func (p *Page) Markdown() (string, error) {
	return HTMLToMarkdown(p.HTML)
}

// HTMLToMarkdown 把 HTML 片段转换为 Markdown
func HTMLToMarkdown(rawHTML string) (string, error) {
	if strings.TrimSpace(rawHTML) == "" {
		return "", nil
	}

	md, err := convertHTML(rawHTML)
	if err != nil {
		return "", fmt.Errorf("转换 Markdown 失败: %w", err)
	}

	return md, nil
}
