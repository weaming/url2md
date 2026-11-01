package url2md

import (
	"regexp"
	"strings"
	"time"
)

// CST 微信公众号展示时间固定为北京时间
var CST = time.FixedZone("CST", 8*3600)

var (
	wechatCreateTime = regexp.MustCompile(`create_time:\s*'(\d{4}-\d{2}-\d{2} \d{2}:\d{2})'`)
	wechatTimestamp  = regexp.MustCompile(`var\s+ct\s*=\s*"(\d{9,11})"`)
	wechatModifyTime = regexp.MustCompile(`modify_time:\s*'(\d{9,11})'`)
	wechatProvince   = regexp.MustCompile(`provinceName:\s*'([^']*)'`)
	wechatCountry    = regexp.MustCompile(`countryName:\s*'([^']*)'`)
	wechatSourceURL  = regexp.MustCompile(`var\s+msg_source_url\s*=\s*'([^']*)'`)
	wechatSourceURL2 = regexp.MustCompile(`source_url:\s*'([^']*)'`)
)

// enrichWechat 补充微信由 JS 渲染的发布时间、修改时间、IP 属地与原文链接
func enrichWechat(rawHTML string, page *Page) {
	if page.PublishTime == "" {
		page.PublishTime = wechatPublishTime(rawHTML)
	}
	if page.ModifiedTime == "" {
		page.ModifiedTime = wechatModifiedTime(rawHTML)
	}
	if page.IPLocation == "" {
		page.IPLocation = wechatIPLocation(rawHTML)
	}
	if page.SourceURL == "" {
		page.SourceURL = wechatSourceLink(rawHTML)
	}
}

// wechatPublishTime 从脚本的 create_time 或 ct 时间戳还原发布时间
func wechatPublishTime(rawHTML string) string {
	if m := wechatCreateTime.FindStringSubmatch(rawHTML); m != nil {
		if t, err := time.ParseInLocation("2006-01-02 15:04", m[1], CST); err == nil {
			return t.Format("2006年01月02日 15:04")
		}
	}

	if m := wechatTimestamp.FindStringSubmatch(rawHTML); m != nil {
		return formatTimestamp(m[1], "2006年01月02日 15:04")
	}

	return ""
}

// wechatModifiedTime 从 modify_time 时间戳还原「修改于」日期
func wechatModifiedTime(rawHTML string) string {
	if m := wechatModifyTime.FindStringSubmatch(rawHTML); m != nil {
		return formatTimestamp(m[1], "2006年01月02日")
	}
	return ""
}

// wechatSourceLink 取「阅读原文」链接
func wechatSourceLink(rawHTML string) string {
	for _, re := range []*regexp.Regexp{wechatSourceURL, wechatSourceURL2} {
		if m := re.FindStringSubmatch(rawHTML); m != nil {
			if link := strings.TrimSpace(m[1]); link != "" {
				return link
			}
		}
	}
	return ""
}

// wechatIPLocation 从 ip_wording 配置里取省份，缺失时退回国家
func wechatIPLocation(rawHTML string) string {
	if m := wechatProvince.FindStringSubmatch(rawHTML); m != nil && cleanText(m[1]) != "" {
		return cleanText(m[1])
	}
	if m := wechatCountry.FindStringSubmatch(rawHTML); m != nil {
		return cleanText(m[1])
	}
	return ""
}

// formatTimestamp 把秒级时间戳按北京时间格式化
func formatTimestamp(seconds string, layout string) string {
	var value int64
	for _, ch := range seconds {
		value = value*10 + int64(ch-'0')
	}
	return time.Unix(value, 0).In(CST).Format(layout)
}
