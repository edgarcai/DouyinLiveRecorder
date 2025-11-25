package spider

type Spider interface {
	GetStreamUrl(url string) (string, error)
	SetProxy(proxyUrl string)
	SetCookies(cookies string)
}
