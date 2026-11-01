package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	rawMode := flag.Bool("raw", false, "输出原始 Markdown 内容")
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
		os.Stdout.WriteString(md)
		return
	}

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
