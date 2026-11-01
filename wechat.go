package url2md

import (
	"regexp"
	"time"
)

// CST 微信公众号展示时间固定为北京时间
var CST = time.FixedZone("CST", 8*3600)

var (
	wechatCreateTime = regexp.MustCompile(`create_time:\s*'(\d{4}-\d{2}-\d{2} \d{2}:\d{2})'`)
	wechatTimestamp  = regexp.MustCompile(`var\s+ct\s*=\s*"(\d{9,11})"`)
	wechatProvince   = regexp.MustCompile(`provinceName:\s*'([^']*)'`)
	wechatCountry    = regexp.MustCompile(`countryName:\s*'([^']*)'`)
)

// enrichWechat 补充微信由 JS 渲染的发布时间与 IP 属地
func enrichWechat(rawHTML string, page *Page) {
	if page.PublishTime == "" {
		page.PublishTime = wechatPublishTime(rawHTML)
	}
	if page.IPLocation == "" {
		page.IPLocation = wechatIPLocation(rawHTML)
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
		var seconds int64
		for _, ch := range m[1] {
			seconds = seconds*10 + int64(ch-'0')
		}
		return time.Unix(seconds, 0).In(CST).Format("2006年01月02日 15:04")
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
