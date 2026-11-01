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
