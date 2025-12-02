package spider

import (
	"fmt"
	"strings"
)

// Factory 根据 URL 创建爬虫
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

// InitSpiders 注册所有可用的爬虫
func InitSpiders() {
	Register("douyin.com", func() Spider { return NewDouyinSpider() })
	Register("tiktok.com", func() Spider { return NewTikTokSpider() })
	Register("kuaishou.com", func() Spider { return NewKuaishouSpider() })
	Register("huya.com", func() Spider { return NewHuyaSpider() })
	Register("douyu.com", func() Spider { return NewDouyuSpider() })
	Register("bilibili.com", func() Spider { return NewBilibiliSpider() })
	Register("youtube.com", func() Spider { return NewYoutubeSpider() })
	Register("youtu.be", func() Spider { return NewYoutubeSpider() })
	Register("twitch.tv", func() Spider { return NewTwitchSpider() })
	Register("xiaohongshu.com", func() Spider { return NewXiaohongshuSpider() })
	Register("xhslink.com", func() Spider { return NewXiaohongshuSpider() })
	Register("yy.com", func() Spider { return NewYYSpider() })
	Register("cc.163.com", func() Spider { return NewNeteaseSpider() })
	Register("bigo.tv", func() Spider { return NewBigoSpider() })
	Register("blued.cn", func() Spider { return NewBluedSpider() })
	Register("qiandurebo.com", func() Spider { return NewQianduSpider() })
	Register("fm.missevan.com", func() Spider { return NewMaoerSpider() })
	Register("look.163.com", func() Spider { return NewLookSpider() })
	Register("twitcasting.tv", func() Spider { return NewTwitCastingSpider() })
	Register("live.baidu.com", func() Spider { return NewBaiduSpider() })
	Register("weibo.com", func() Spider { return NewWeiboSpider() })
	Register("fanxing.kugou.com", func() Spider { return NewKugouSpider() })
	Register("fanxing2.kugou.com", func() Spider { return NewKugouSpider() })
	Register("huodong.m.taobao.com", func() Spider { return NewTaobaoSpider() })
	Register("lives.jd.com", func() Spider { return NewJDSpider() })
	Register("live.shopee.sg", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.com.my", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.co.th", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.co.id", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.vn", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.ph", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.tw", func() Spider { return NewShopeeSpider() })
	Register("live.shopee.com.br", func() Spider { return NewShopeeSpider() })
	Register("zhihu.com", func() Spider { return NewZhihuSpider() })
	Register("live.acfun.cn", func() Spider { return NewAcfunSpider() })
	Register("liveme.com", func() Spider { return NewLiveMeSpider() })
	Register("showroom-live.com", func() Spider { return NewShowRoomSpider() })
	Register("chzzk.naver.com", func() Spider { return NewCHZZKSpider() })
	Register("faceit.com", func() Spider { return NewFaceitSpider() })
	Register("picarto.tv", func() Spider { return NewPicartoSpider() })
	Register("17.live", func() Spider { return NewLive17Spider() })
	Register("lang.live", func() Spider { return NewLangliveSpider() })
	Register("huajiao.com", func() Spider { return NewHuajiaoSpider() })
	Register("7u66.com", func() Spider { return NewLiuxingSpider() })
	Register("tlclw.com", func() Spider { return NewChangliaoSpider() })
	Register("ybw1666.com", func() Spider { return NewYinboSpider() })
	Register("inke.cn", func() Spider { return NewYingkeSpider() })
	Register("haixiutv.com", func() Spider { return NewHaixiuSpider() })
	Register("lehaitv.com", func() Spider { return NewHaixiuSpider() }) // Lehaitv uses HaixiuSpider
	Register("vvxqiu.com", func() Spider { return NewVVxqiuSpider() })
	Register("pp.weimipopo.com", func() Spider { return NewPpliveSpider() })
	Register("catshow168.com", func() Spider { return NewPpliveSpider() }) // Huamao uses PpliveSpider
	Register("6.cn", func() Spider { return NewRoom6Spider() })
	Register("miguvideo.com", func() Spider { return NewMiguSpider() })
	Register("lailianjie.com", func() Spider { return NewLianjieSpider() })
	Register("laixiu.com", func() Spider { return NewLaixiuSpider() })
}
