package url2md

import (
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	ogTitle       = "og:title"
	ogDescription = "og:description"
	ogSiteName    = "og:site_name"
	OG_IMAGE      = "og:image"
	ogKeywords    = "og:keywords"
	ogImageWidth  = "og:image:width"
	ogImageHeight = "og:image:height"
	ogAuthor      = "og:article:author"
	articleTime   = "article:published_time"
)

var spaceRegexp = regexp.MustCompile(`\s+`)

// fillMeta 提取 og 元信息与 favicon
func (p *Page) fillMeta(doc *goquery.Document, baseURL *url.URL) {
	doc.Find(`meta[property^="og:"]`).Each(func(_ int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		switch property {
		case ogTitle:
			p.Title = cleanText(content)
		case ogDescription:
			p.Description = cleanText(content)
		case ogSiteName:
			p.SiteName = cleanText(content)
		case OG_IMAGE:
			p.ImageURL = content
		case ogKeywords:
			p.Keywords = splitKeywords(content)
		case ogAuthor:
			p.Author = cleanText(content)
		case articleTime:
			p.PublishTime = cleanText(content)
		case ogImageWidth:
			p.ImageWidth, _ = strconv.Atoi(strings.TrimSpace(content))
		case ogImageHeight:
			p.ImageHeight, _ = strconv.Atoi(strings.TrimSpace(content))
		}
	})

	if p.Title == "" {
		p.Title = cleanText(doc.Find("title").First().Text())
	}

	p.ImageURL = resolveURL(baseURL, p.ImageURL)

	doc.Find(`link[rel="icon"], link[rel="shortcut icon"], link[rel="apple-touch-icon"]`).
		First().
		Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			p.IconURL = resolveURL(baseURL, href)
		})
}

// cleanText 反转义 HTML 实体并压缩空白
func cleanText(text string) string {
	text = html.UnescapeString(text)
	return strings.TrimSpace(spaceRegexp.ReplaceAllString(text, " "))
}

// splitKeywords 拆分 og:keywords 为关键词列表
func splitKeywords(content string) []string {
	parts := strings.Split(content, ",")
	keywords := make([]string, 0, len(parts))
	for _, part := range parts {
		if keyword := cleanText(part); keyword != "" {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

// resolveURL 把相对链接补全为绝对链接
func resolveURL(baseURL *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || baseURL == nil {
		return ref
	}

	parsed, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	return baseURL.ResolveReference(parsed).String()
}
