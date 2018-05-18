package platform

type Douyu struct {
	roomId    int
	originUrl string
}

var douyuClient Douyu

func New(url string) {
	if douyuClient.originUrl == "" {
		douyuClient = Douyu{
			originUrl: url,
		}
	}
}
