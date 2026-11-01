package url2md

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// LONG_HEADING_MAX 超过该长度的标题降级为段落，微信常把正文段落写成 h1
const LONG_HEADING_MAX = 40

// siteRule 站点级的正文提取规则，命中后优先于通用 readability 提取
type siteRule struct {
	// Hosts 站点域名，后缀匹配
	Hosts []string
	// Title 标题选择器，按顺序取第一个非空文本
	Title []string
	// Author 作者选择器，按顺序取第一个非空文本
	Author []string
	// SiteName 站点/账号名选择器
	SiteName []string
	// PublishTime 发布时间选择器
	PublishTime []string
	// Location IP 属地等地域信息选择器
	Location []string
	// Labels 原创、转载等标签选择器
	Labels []string
	// Content 正文容器选择器
	Content string
	// Drop 需要从正文里移除的节点选择器
	Drop []string
	// Enrich 站点特有的脚本变量补充（JS 渲染的字段，如发布时间、IP 属地）
	Enrich func(rawHTML string, page *Page)
}

var siteRules = []siteRule{
	{
		Hosts:       []string{"mp.weixin.qq.com"},
		Title:       []string{"#activity-name", "h1.rich_media_title"},
		Author:      []string{"#js_author_name", "#js_author_name_text", ".rich_media_meta_text"},
		SiteName:    []string{"#js_name", ".rich_media_meta_nickname"},
		PublishTime: []string{"#publish_time"},
		Location:    []string{"#js_ip_wording"},
		Labels:      []string{"#copyright_logo"},
		Content:     "#js_content",
		Drop:        []string{"script", "style", "noscript", "iframe", "svg", "mp-ad", "mp-common-profile", "qqmusic"},
		Enrich:      enrichWechat,
	},
}

// findRule 按域名查找站点规则
func findRule(u *url.URL) *siteRule {
	if u == nil {
		return nil
	}

	host := strings.ToLower(u.Hostname())
	for i := range siteRules {
		for _, want := range siteRules[i].Hosts {
			if host == want || strings.HasSuffix(host, "."+want) {
				return &siteRules[i]
			}
		}
	}
	return nil
}

// apply 按规则填充元信息与正文，返回值表示规则是否命中
func (r *siteRule) apply(doc *goquery.Document, rawHTML string, page *Page) bool {
	content := doc.Find(r.Content).First()
	if content.Length() == 0 {
		return false
	}

	for _, selector := range r.Drop {
		content.Find(selector).Remove()
	}

	// 懒加载图片：data-src 才是真实地址
	content.Find("img").Each(func(_ int, s *goquery.Selection) {
		if src, ok := s.Attr("data-src"); ok && strings.TrimSpace(src) != "" {
			s.SetAttr("src", src)
		}
		s.RemoveAttr("data-src")
	})

	demoteLongHeadings(content)

	text := nodeText(content)
	if text == "" {
		return false
	}

	contentHTML, err := content.Html()
	if err != nil {
		return false
	}

	page.HTML = contentHTML
	page.Text = text

	if title := firstText(doc, r.Title); title != "" {
		page.Title = title
	}
	if author := firstText(doc, r.Author); author != "" {
		page.Author = author
	}
	if siteName := firstText(doc, r.SiteName); siteName != "" {
		page.SiteName = siteName
	}
	if publishTime := firstText(doc, r.PublishTime); publishTime != "" {
		page.PublishTime = publishTime
	}
	if location := firstText(doc, r.Location); location != "" {
		page.IPLocation = location
	}
	if labels := allTexts(doc, r.Labels); len(labels) > 0 {
		page.Labels = labels
	}

	if r.Enrich != nil {
		r.Enrich(rawHTML, page)
	}

	return true
}

// demoteLongHeadings 把过长的标题降级为段落，避免正文段落变成 Markdown 标题
func demoteLongHeadings(sel *goquery.Selection) {
	sel.Find("h1,h2,h3,h4,h5,h6").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Text())
		if utf8.RuneCountInString(label) <= LONG_HEADING_MAX {
			return
		}

		inner, err := s.Html()
		if err != nil {
			return
		}
		s.ReplaceWithHtml("<p>" + inner + "</p>")
	})
}

// firstText 返回第一个选择器命中的非空文本，meta 标签取 content 属性
func firstText(doc *goquery.Document, selectors []string) string {
	for _, selector := range selectors {
		node := doc.Find(selector).First()
		if node.Length() == 0 {
			continue
		}

		text := ""
		if content, ok := node.Attr("content"); ok && node.Is("meta") {
			text = content
		} else {
			text = node.Text()
		}

		if text = cleanText(text); text != "" {
			return text
		}
	}
	return ""
}

// allTexts 返回所有选择器命中的非空文本
func allTexts(doc *goquery.Document, selectors []string) []string {
	texts := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			if text := cleanText(s.Text()); text != "" {
				texts = append(texts, text)
			}
		})
	}
	return texts
}

// nodeText 把节点内容转换为按块分行的纯文本
func nodeText(sel *goquery.Selection) string {
	node := sel.Clone()
	node.Find("br").ReplaceWithHtml("\n")
	node.Find("p,div,section,h1,h2,h3,h4,h5,h6,li,tr,blockquote,pre,figure,figcaption,table,img").
		AfterHtml("\n")

	lines := make([]string, 0, 64)
	for _, line := range strings.Split(node.Text(), "\n") {
		if line = cleanText(line); line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}
