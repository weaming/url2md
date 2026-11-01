package url2md

import (
	"fmt"
	"strings"
)

// Page 表示一个抓取并解析完成的网页
type Page struct {
	URL string

	Title        string
	Author       string
	SiteName     string
	PublishTime  string
	ModifiedTime string
	IPLocation   string
	Labels       []string
	SourceURL    string
	Description  string
	Keywords     []string
	ImageURL     string
	ImageWidth   int
	ImageHeight  int
	IconURL      string

	// Text 由 readability 提取的正文纯文本
	Text string
	// HTML 由 readability 提取的正文 HTML
	HTML string
}

// Markdown 把正文 HTML 转换为 Markdown
func (p *Page) Markdown() (string, error) {
	return HTMLToMarkdown(p.HTML)
}

// FrontMatter 返回 YAML front matter 形式的元信息，空字段自动省略
func (p *Page) FrontMatter() string {
	var b strings.Builder
	b.WriteString("---\n")

	writeYAMLField(&b, "title", p.Title)
	writeYAMLField(&b, "url", p.URL)
	writeYAMLField(&b, "author", p.Author)
	writeYAMLField(&b, "account", p.SiteName)
	writeYAMLField(&b, "published", p.PublishTime)
	writeYAMLField(&b, "modified", p.ModifiedTime)
	writeYAMLField(&b, "ip_location", p.IPLocation)
	writeYAMLField(&b, "description", p.Description)
	writeYAMLField(&b, "image", p.ImageURL)
	writeYAMLField(&b, "source_url", p.SourceURL)
	writeYAMLRaw(&b, "labels", yamlList(p.Labels))
	writeYAMLRaw(&b, "keywords", yamlList(p.Keywords))

	b.WriteString("---\n\n")
	return b.String()
}

// writeYAMLField 写入单个 YAML 字段，空值跳过
func writeYAMLField(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "%s: %s\n", key, yamlQuote(value))
}

// writeYAMLRaw 写入已格式化好的 YAML 值
func writeYAMLRaw(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "%s: %s\n", key, value)
}

// yamlList 生成 YAML 行内数组，空列表返回空字符串
func yamlList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			quoted = append(quoted, yamlQuote(item))
		}
	}
	if len(quoted) == 0 {
		return ""
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// yamlQuote 用双引号包裹并转义，避免标点破坏 YAML 结构
func yamlQuote(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	return `"` + escaped + `"`
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
