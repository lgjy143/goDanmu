package platform

import (
	"danmu/net"
	"danmu/utils/log"
	"fmt"
	"net/url"
	"regexp"

	"github.com/bitly/go-simplejson"
)

type DouyuClient struct {
	roomId     int
	roomName   string
	showId     int
	ownerUid   int
	ownerName  string
	serverIp   string
	serverPort string
	originUrl  string
}

var douyuClient DouyuClient

func Douyu(url string) {
	if douyuClient.originUrl == "" {
		fmt.Println(douyuClient.originUrl)
		douyuClient = DouyuClient{
			originUrl: url,
			roomId:    0,
		}
		douyuClient.getClientInfo()
	}

	fmt.Println(douyuClient)

}

/**
 * 获取斗鱼房间信息
 */
func (client *DouyuClient) getClientInfo() {
	pageHTML := net.Get(client.originUrl)
	reRoom := regexp.MustCompile(`var\s\$ROOM\s=\s({.*});`)
	roomStr := reRoom.FindStringSubmatch(pageHTML)[1]
	roomJSON, err := simplejson.NewJson([]byte(roomStr))

	if err != nil {
		log.Fatal(err.Error())
	}

	client.roomId, _ = roomJSON.Get("room_id").Int()
	client.roomName, _ = roomJSON.Get("room_name").String()
	client.showId, _ = roomJSON.Get("show_id").Int()
	client.ownerUid, _ = roomJSON.Get("owner_uid").Int()
	client.ownerName, _ = roomJSON.Get("owner_name").String()

	reAuth := regexp.MustCompile(`\$ROOM\.args\s=\s({.*});`)
	authStr := reAuth.FindStringSubmatch(pageHTML)[1]
	authJSON, err := simplejson.NewJson([]byte(authStr))
	if err != nil {
		log.Fatal(err.Error())
	}

	serversStr, _ := authJSON.Get("server_config").String()
	serversStr, _ = url.QueryUnescape(serversStr)
	serversJSON, err := simplejson.NewJson([]byte(serversStr))
	if err != nil {
		log.Fatal(err.Error())
	}
	server := serversJSON.GetIndex(0)
	client.serverIp, _ = server.Get("ip").String()
	client.serverPort, _ = server.Get("port").String()
}
