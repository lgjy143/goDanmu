package platform

import (
	"danmu/utils/log"
	"fmt"
)

type DouyuClient struct {
	roomId    int
	originUrl string
}

var douyuClient DouyuClient

func Douyu(url string) {
	log.Info(url)

	if douyuClient.originUrl == "" {
		fmt.Println(douyuClient.originUrl)
		douyuClient = DouyuClient{
			originUrl: url,
			roomId:    0,
		}
		douyuClient.getRoomId()
	}

}

/**
 * 获取斗鱼RoomID
 */
func (client *DouyuClient) getRoomId() {
	// pageHtml := net.Get(client.originUrl)

}
