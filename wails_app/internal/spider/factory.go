package spider

import (
	"fmt"
	"strings"
)

// Factory creates spiders based on URL
type Factory struct {
	spiders map[string]func() Spider
}

var GlobalFactory = &Factory{
	spiders: make(map[string]func() Spider),
}

func Register(domain string, constructor func() Spider) {
	GlobalFactory.spiders[domain] = constructor
}

func GetSpider(url string) (Spider, error) {
	for domain, constructor := range GlobalFactory.spiders {
		if strings.Contains(url, domain) {
			return constructor(), nil
		}
	}
	return nil, fmt.Errorf("no spider found for url: %s", url)
}

// Initialize registers all available spiders
func InitSpiders() {
	Register("douyin.com", func() Spider { return NewDouyinSpider() })
	Register("tiktok.com", func() Spider { return NewTikTokSpider() })
	Register("kuaishou.com", func() Spider { return NewKuaishouSpider() })
	Register("huya.com", func() Spider { return NewHuyaSpider() })
	Register("douyu.com", func() Spider { return NewDouyuSpider() })
	Register("bilibili.com", func() Spider { return NewBilibiliSpider() })
}
