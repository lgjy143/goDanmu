package douyu

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
	"time"

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
	conn       net.Conn
	wg         sync.WaitGroup
	closeFlag  chan bool
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

	if err := douyuClient.Connect(); err != nil {
		log.Fatal("connect danmu server error")
	}

	fmt.Println(douyuClient)

}

/**
 * 获取斗鱼房间信息
 */
func (client *DouyuClient) getClientInfo() {
	pageHTML := utils.Get(client.originUrl)
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

/**
 * 与斗鱼弹幕服务器建立连接
 */
func (client *DouyuClient) Connect() error {
	var danmuServerStr = "openbarrage.douyutv.com:8601"
	conn, err := net.Dial("tcp", danmuServerStr)
	if err != nil {
		return err
	}

	client.conn = conn
	// join Room
	client.joinRoom()
	client.wg.Add(2)
	// heart
	go client.heartbeat()
	// chatMsg
	go client.chatMsg()
	client.wg.Wait()
	log.Infof("%s connected.", danmuServerStr)
	return nil
}

/**
 * Read data from connection and process
 */
func (client *DouyuClient) ReceiveMsg() ([]byte, int, error) {
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

/**
 * 心跳检测
 */
func (client *DouyuClient) heartbeat() {
	defer client.wg.Done()
	tick := time.Tick(45 * time.Second)
loop:
	for {
		select {
		case _, ok := <-client.closeFlag:
			if !ok {
				break loop
			}
		case <-tick:
			heartbeatMsg := NewMessage(nil, MESSAGE_TO_SERVER).
				SetField("type", "keeplive").
				SetField("tick", time.Now().Unix())
			fmt.Println("heart")
			_, err := client.conn.Write(heartbeatMsg.Encode())
			if err != nil {
				log.Error("heartbeat failed, " + err.Error())
			}
		}
	}
}

func (client *DouyuClient) chatMsg() {
	defer client.wg.Done()
loop:
	for {
		select {
		case _, ok := <-client.closeFlag:
			if !ok {
				break loop
			}
		default:
			b, code, err := client.ReceiveMsg()
			if err != nil {
				log.Error(err, code)
				break loop
			}

			// analize message
			msg := NewMessage(nil, MESSAGE_FROM_SERVER).Decode(b, code)
			if msg.GetStringField("type") == "chatmsg" {
				log.Infof("type %s, content %s", msg.GetStringField("type"), msg.GetStringField("txt"))
			}
		}
	}
}

/**
 * 加入斗鱼弹幕房间
 */
func (client *DouyuClient) joinRoom() error {
	var room = client.roomId
	loginMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", MSG_TYPE_LOGINREQ).
		SetField("roomid", room)

	log.Infof("joining room %d...", room)
	if _, err := client.conn.Write(loginMessage.Encode()); err != nil {
		return err
	}

	b, code, err := client.ReceiveMsg()
	if err != nil {
		return err
	}

	// TODO assert(code == MESSAGE_FROM_SERVER)
	log.Infof("room %d joined", room)
	loginRes := NewMessage(nil, MESSAGE_FROM_SERVER).Decode(b, code)
	// isLive := loginRes.GetStringField("live_stat")
	// if isLive == 0 {
	log.Infof("room %d live status %s", room, loginRes.GetStringField("live_stat"))
	// }

	joinMessage := NewMessage(nil, MESSAGE_TO_SERVER).
		SetField("type", "joingroup").
		SetField("rid", room).
		SetField("gid", "-9999")

	log.Infof("joining group %d...", -9999)
	_, err = client.conn.Write(joinMessage.Encode())
	if err != nil {
		return err
	}
	log.Infof("group %d joined", -9999)
	return nil
}
