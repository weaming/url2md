# url2md

抓取网页、提取正文并转换为 Markdown 的 Go 库，附带命令行工具。

## 库

```go
page, err := url2md.Fetch("https://example.com/post")
md, err := page.Markdown() // 正文 Markdown
fmt.Println(page.Title, page.Text, page.HTML)
```

- `Fetch(url)`：请求网页并解析，返回 `*Page`（og 元信息 + readability 正文）
- `FetchMarkdown(url)`：一步拿到正文 Markdown
- `Parse(html, url)`：不联网，从已有 HTML 解析
- `HTMLToMarkdown(html)`：把 HTML 片段转 Markdown

正文提取以站点规则（`rules.go`）优先，未命中站点时退回 go-readability。
目前内置：微信公众号 `mp.weixin.qq.com`（`#js_content` 正文、懒加载图片、标题/作者/公众号/发布时间/IP 属地/原创标签）。

`Page.FrontMatter()` 输出 YAML 元信息（空字段省略），`Page.Markdown()` 输出正文。

`Page` 字段：`Title`、`Author`、`SiteName`（公众号）、`PublishTime`、`ModifiedTime`、
`IPLocation`、`Labels`（原创/转载）、`SourceURL`（阅读原文）、`Description`、`Keywords`、
`ImageURL`、`IconURL`、`Text`、`HTML`。

微信的发布时间、修改时间、IP 属地与原文链接由页面脚本渲染，`wechat.go` 从
`create_time` / `ct` / `modify_time` 时间戳、`ip_wording` 配置与 `msg_source_url` 还原。

## 命令行

```bash
go install github.com/weaming/url2md/cmd/url2md@latest

url2md https://example.com/post      # 终端渲染
url2md -raw https://example.com/post # 输出 YAML front matter + 原始 Markdown
```
