package spider

type StreamInfo struct {
	Url        string
	Title      string
	AnchorName string
}

type Spider interface {
	GetStreamUrl(url string) (*StreamInfo, error)
	SetProxy(proxyUrl string)
	SetCookies(cookies string)
}
