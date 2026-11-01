package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/weaming/url2md"
)

func init() {
	flag.Usage = func() {
		fmt.Println("用法: url2md [选项] <url>")
		fmt.Println("将 URL 的内容转换为 Markdown 格式输出")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func main() {
	rawMode := flag.Bool("raw", false, "输出原始 Markdown（正文前附 YAML front matter）")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	page, err := url2md.Fetch(flag.Arg(0))
	if err != nil {
		log.Fatalf("获取内容失败: %v", err)
	}

	md, err := page.Markdown()
	if err != nil {
		log.Fatalf("转换 Markdown 失败: %v", err)
	}

	if *rawMode {
		os.Stdout.WriteString(page.FrontMatter())
		os.Stdout.WriteString(md)
		return
	}

	printMeta(page)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		log.Fatalf("创建渲染器失败: %v", err)
	}

	rendered, err := renderer.Render(md)
	if err != nil {
		log.Fatalf("渲染 Markdown 失败: %v", err)
	}

	fmt.Print(rendered)
}

// printMeta 在终端渲染前打印标题与元信息
func printMeta(page *url2md.Page) {
	fmt.Println(page.Title)

	parts := make([]string, 0, 6)
	for _, value := range []string{page.Author, page.SiteName, page.PublishTime, page.IPLocation} {
		if value != "" {
			parts = append(parts, value)
		}
	}
	parts = append(parts, page.Labels...)
	if len(parts) > 0 {
		fmt.Println(strings.Join(parts, " · "))
	}

	fmt.Println(page.URL)
	fmt.Println()
}
