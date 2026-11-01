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
	// Content 正文容器选择器
	Content string
	// Drop 需要从正文里移除的节点选择器
	Drop []string
}

var siteRules = []siteRule{
	{
		Hosts:   []string{"mp.weixin.qq.com"},
		Title:   []string{"#activity-name", "h1.rich_media_title"},
		Author:  []string{"#js_name", ".rich_media_meta_nickname"},
		Content: "#js_content",
		Drop:    []string{"script", "style", "noscript", "iframe", "svg", "mp-ad", "mp-common-profile", "qqmusic"},
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

// extract 按照规则提取正文 HTML 与纯文本，返回空字符串表示规则未命中
func (r *siteRule) extract(doc *goquery.Document) (string, string) {
	content := doc.Find(r.Content).First()
	if content.Length() == 0 {
		return "", ""
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
		return "", ""
	}

	rawHTML, err := content.Html()
	if err != nil {
		return "", ""
	}

	return rawHTML, text
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
