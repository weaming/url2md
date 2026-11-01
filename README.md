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

## 命令行

```bash
go install github.com/weaming/url2md/cmd/url2md@latest

url2md https://example.com/post      # 终端渲染
url2md -raw https://example.com/post # 输出原始 Markdown
```
