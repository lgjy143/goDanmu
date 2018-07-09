package bilibili

import (
	"danmu/utils"
	"danmu/utils/log"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"sync"

	"github.com/bitly/go-simplejson"
)

type BiliClient struct {
	roomId     int
	roomName   string
	showId     int
	ownerUid   int
	ownerName  string
	serverIp   string
	serverPort string
	originUrl  string
	conn       net.Conn
	wg         sync.WaitGroup
	closeFlag  chan bool
}

var douyuClient BiliClient

func Douyu(url string) {
	if douyuClient.originUrl == "" {
		fmt.Println(douyuClient.originUrl)
		douyuClient = BiliClient{
			originUrl: url,
			roomId:    0,
		}
		douyuClient.getClientInfo(288016)
	}

	if err := douyuClient.Connect(); err != nil {
		log.Fatal("connect danmu server error")
	}

	fmt.Println(douyuClient)

}

/**
 * 获取斗鱼房间信息
 */
func (client *BiliClient) getClientInfo(roomId int) {
	pageHTML := utils.Get(client.originUrl)
	reRoom := regexp.MustCompile(`var\s\$ROOM\s=\s({.*});`)
	if roomArr := reRoom.FindStringSubmatch(pageHTML); len(roomArr) < 2 {
		client.roomId = roomId
		return
	}

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

/**
 * 与斗鱼弹幕服务器建立连接
 */
func (client *BiliClient) Connect() error {
	var danmuServerStr = "openbarrage.douyutv.com:8601"
	conn, err := net.Dial("tcp", danmuServerStr)
	if err != nil {
		return err
	}

	client.conn = conn
	// join Room
	// client.joinRoom()
	client.wg.Add(2)
	// heart
	// chatMsg
	client.wg.Wait()
	log.Infof("%s connected.", danmuServerStr)
	return nil
}

/**
 * Read data from connection and process
 */
func (client *BiliClient) ReceiveMsg() ([]byte, int, error) {
	buf := make([]byte, 512)
	if _, err := io.ReadFull(client.conn, buf[:12]); err != nil {
		return buf, 0, err
	}

	// 12 bytes header
	// 4byte for packet length
	pl := binary.LittleEndian.Uint32(buf[:4])

	// ignore buf[4:8]

	// 2byte for message type
	code := binary.LittleEndian.Uint16(buf[8:10])

	// 1byte for secret
	// 1byte for reserved

	// body content length(include ENDING)
	cl := pl - 8

	if cl > 512 {
		// expand buffer
		buf = make([]byte, cl)
	}
	if _, err := io.ReadFull(client.conn, buf[:cl]); err != nil {
		return buf, int(code), err
	}
	// exclude ENDING
	return buf[:cl-1], int(code), nil
}
