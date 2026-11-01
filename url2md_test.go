package url2md

import (
	"strings"
	"testing"
)

const sampleHTML = `<!DOCTYPE html>
<html>
<head>
	<title>页面标题</title>
	<meta property="og:title" content="OG &amp; Title">
	<meta property="og:description" content="一段   描述">
	<meta property="og:site_name" content="Example Site">
	<meta property="og:keywords" content="go, markdown , ,  html">
	<meta property="og:image" content="/images/cover.png">
	<meta property="og:image:width" content="1200">
	<meta property="og:image:height" content="630">
	<link rel="icon" href="/favicon.ico">
</head>
<body>
	<article>
		<h1>Hello World</h1>
		<p>这是<strong>正文</strong>段落。</p>
		<ul><li>item one</li><li>item two</li></ul>
	</article>
</body>
</html>`

func TestParseMeta(t *testing.T) {
	page, err := Parse(sampleHTML, "https://example.com/posts/1")
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}

	if page.Title != "OG & Title" {
		t.Errorf("Title = %q", page.Title)
	}
	if page.Description != "一段 描述" {
		t.Errorf("Description = %q", page.Description)
	}
	if page.SiteName != "Example Site" {
		t.Errorf("SiteName = %q", page.SiteName)
	}
	if page.ImageURL != "https://example.com/images/cover.png" {
		t.Errorf("ImageURL = %q", page.ImageURL)
	}
	if page.IconURL != "https://example.com/favicon.ico" {
		t.Errorf("IconURL = %q", page.IconURL)
	}
	if page.ImageWidth != 1200 || page.ImageHeight != 630 {
		t.Errorf("图片尺寸 = %dx%d", page.ImageWidth, page.ImageHeight)
	}
	if len(page.Keywords) != 3 || page.Keywords[0] != "go" || page.Keywords[2] != "html" {
		t.Errorf("Keywords = %v", page.Keywords)
	}
}

func TestParseTitleFallback(t *testing.T) {
	rawHTML := `<html><head><title>  兜底
标题  </title></head><body><p>body</p></body></html>`

	page, err := Parse(rawHTML, "https://example.com")
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	if page.Title != "兜底 标题" {
		t.Errorf("Title = %q", page.Title)
	}
}

func TestMarkdown(t *testing.T) {
	md, err := HTMLToMarkdown(`<h1>标题</h1><p>正文 <strong>加粗</strong></p><ul><li>a</li><li>b</li></ul>`)
	if err != nil {
		t.Fatalf("HTMLToMarkdown 失败: %v", err)
	}

	for _, want := range []string{"# 标题", "**加粗**", "- a", "- b"} {
		if !strings.Contains(md, want) {
			t.Errorf("Markdown 缺少 %q，实际输出:\n%s", want, md)
		}
	}
}

func TestHTMLToMarkdownEmpty(t *testing.T) {
	md, err := HTMLToMarkdown("   ")
	if err != nil {
		t.Fatalf("HTMLToMarkdown 失败: %v", err)
	}
	if md != "" {
		t.Errorf("期望空输出，实际 %q", md)
	}
}

const wechatHTML = `<!DOCTYPE html>
<html>
<head>
	<title>页面标题</title>
	<meta property="og:title" content="og 标题">
	<meta property="og:article:author" content="og 作者">
</head>
<body>
	<h1 id="activity-name">Super-agent 设计探索</h1>
	<a id="js_name">某某团队</a>
	<span id="copyright_logo">原创</span>
	<span id="js_author_name">某某产品部</span>
	<div id="js_content">
		<p><span>第一段正文，说明背景。</span></p>
		<h1><span>这是一段被写成标题的长正文，长度明显超过四十个字，用来验证标题降级逻辑是否生效，应该被输出成普通段落而不是 Markdown 标题。</span></h1>
		<p><img data-src="https://mmbiz.qpic.cn/sz_mmbiz_jpg/cover/640?wx_fmt=jpeg" src="data:image/svg+xml;charset=utf8,%3Csvg%3E"></p>
		<script>var noise = "不该出现";</script>
	</div>
	<script>
		var ct = "1793449787";
		var msg_source_url = 'https://github.com/example/repo';
		window.ip_wording = { countryName: '中国', provinceName: '浙江' };
		modify_time: '1793467167' * 1,
	</script>
</body>
</html>`

func TestParseWechat(t *testing.T) {
	page, err := Parse(wechatHTML, "https://mp.weixin.qq.com/s/abcdefg")
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}

	if page.Title != "Super-agent 设计探索" {
		t.Errorf("Title = %q", page.Title)
	}
	if page.Author != "某某产品部" {
		t.Errorf("Author = %q", page.Author)
	}
	if page.SiteName != "某某团队" {
		t.Errorf("SiteName = %q", page.SiteName)
	}
	if page.PublishTime != "2026年10月31日 20:29" {
		t.Errorf("PublishTime = %q", page.PublishTime)
	}
	if page.IPLocation != "浙江" {
		t.Errorf("IPLocation = %q", page.IPLocation)
	}
	if page.ModifiedTime != "2026年11月01日" {
		t.Errorf("ModifiedTime = %q", page.ModifiedTime)
	}
	if page.SourceURL != "https://github.com/example/repo" {
		t.Errorf("SourceURL = %q", page.SourceURL)
	}
	if len(page.Labels) != 1 || page.Labels[0] != "原创" {
		t.Errorf("Labels = %v", page.Labels)
	}
	if !strings.Contains(page.Text, "第一段正文") {
		t.Errorf("Text 缺少正文: %q", page.Text)
	}
	if strings.Contains(page.Text, "不该出现") {
		t.Errorf("Text 包含 script 内容: %q", page.Text)
	}

	md, err := page.Markdown()
	if err != nil {
		t.Fatalf("Markdown 失败: %v", err)
	}

	if !strings.Contains(md, "![](https://mmbiz.qpic.cn/sz_mmbiz_jpg/cover/640?wx_fmt=jpeg)") {
		t.Errorf("Markdown 未使用 data-src 图片:\n%s", md)
	}
	if strings.Contains(md, "# 这是一段被写成标题的长正文") {
		t.Errorf("过长的标题未降级:\n%s", md)
	}
	if !strings.Contains(md, "这是一段被写成标题的长正文") {
		t.Errorf("长正文丢失:\n%s", md)
	}
}

func TestParseWechatFallback(t *testing.T) {
	// 缺少 #js_content（如文章被删除）时退回 readability，不应报错
	rawHTML := `<html><head><title>提示</title></head><body><p>该内容已被发布者删除</p></body></html>`

	page, err := Parse(rawHTML, "https://mp.weixin.qq.com/s/abcdefg")
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	if !strings.Contains(page.Text, "该内容已被发布者删除") {
		t.Errorf("Text = %q", page.Text)
	}
}

func TestFrontMatter(t *testing.T) {
	page := &Page{
		URL:         "https://mp.weixin.qq.com/s/abcdefg",
		Title:       `标题: 带 "引号"`,
		Author:      "某某产品部",
		SiteName:    "某某团队",
		PublishTime: "2026年10月31日 20:29",
		IPLocation:  "浙江",
		Labels:      []string{"原创"},
		Keywords:    []string{"a", "b"},
	}

	fm := page.FrontMatter()
	if !strings.HasPrefix(fm, "---\n") || !strings.HasSuffix(fm, "---\n\n") {
		t.Fatalf("front matter 结构不对:\n%s", fm)
	}

	for _, want := range []string{
		`title: "标题: 带 \"引号\""`,
		`published: "2026年10月31日 20:29"`,
		`ip_location: "浙江"`,
		`labels: ["原创"]`,
		`keywords: ["a", "b"]`,
	} {
		if !strings.Contains(fm, want) {
			t.Errorf("front matter 缺少 %q:\n%s", want, fm)
		}
	}
}

func TestFrontMatterSkipsEmpty(t *testing.T) {
	page := &Page{URL: "https://example.com/post", Title: "标题"}
	fm := page.FrontMatter()

	if strings.Contains(fm, "author") || strings.Contains(fm, "labels") {
		t.Errorf("空字段未被省略:\n%s", fm)
	}
}
