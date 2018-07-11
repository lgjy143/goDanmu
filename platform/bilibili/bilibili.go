package bilibili

import (
	"danmu/utils"
	"danmu/utils/log"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
)

type BiliClient struct {
	roomID     int
	roomName   string
	showID     int
	ownerUID   int
	ownerName  string
	serverIP   string
	serverPort string
	originURL  string
	liveStatus int
	conn       net.Conn
	wg         sync.WaitGroup
	closeFlag  chan bool
}

var biliClient BiliClient

func Bilibili(url string) {
	if biliClient.originURL == "" {
		fmt.Println(biliClient.originURL)
		biliClient = BiliClient{
			originURL: url,
			roomID:    0,
			closeFlag: make(chan bool),
		}
		biliClient.getClientInfo()
	}
}

/**
 * 获取B站房间信息
 */
func (client *BiliClient) getClientInfo() {
	var roomInitAddr = "https://api.live.bilibili.com/room/v1/Room/room_init?id=%s"
	reURL := regexp.MustCompile(`.*\/(\d+)$`)
	roomArr := reURL.FindStringSubmatch(client.originURL)
	if len(roomArr) == 2 {
		roomInitHTML := utils.Get(fmt.Sprintf(roomInitAddr, roomArr[1]))

		roomJSON, err := simplejson.NewJson([]byte(roomInitHTML))

		if err != nil {
			log.Fatal(err.Error())
		}

		if resOk, _ := roomJSON.Get("msg").String(); resOk == "ok" {
			roomData := roomJSON.Get("data")
			client.roomID, _ = roomData.Get("room_id").Int()
			client.ownerUID, _ = roomData.Get("uid").Int()
			client.liveStatus, _ = roomData.Get("live_status").Int()
		} else {
			log.Fatal("Room Init Failed")
		}

		log.Info("Get Barrage Servers Info")
		dmIP, dmPort, err := getBarrageServer(client.roomID)
		if err != nil {
			log.Fatal(err)
		}
		client.serverIP = dmIP
		client.serverPort = dmPort

		client.Connect()

	}
}

func getBarrageServer(roomID int) (string, string, error) {
	apiAddr := fmt.Sprintf("http://live.bilibili.com/api/player?id=cid:%d", roomID)
	serverHTML := utils.Get(apiAddr)
	regDmServer := regexp.MustCompile(`<dm_server>(.*)<\/dm_server>`)
	regDmPort := regexp.MustCompile(`<dm_port>(\d+)<\/dm_port>`)

	dmServer := regDmServer.FindStringSubmatch(serverHTML)
	dmPort := regDmPort.FindStringSubmatch(serverHTML)

	if len(dmServer) == 2 && len(dmPort) == 2 {
		return dmServer[1], dmPort[1], nil
	}
	return "", "", errors.New("Get Barrage Server Error")
}

/**
 * 与B站弹幕服务器建立连接
 */
func (client *BiliClient) Connect() error {
	var danmuServerStr = client.serverIP + ":" + client.serverPort
	conn, err := net.DialTimeout("tcp", danmuServerStr, 5*time.Second)
	if err != nil {
		return err
	}
	log.Info("Connect Success")
	client.conn = conn
	// hand shake
	client.conn.Write(NewHandshakeMessage(client.roomID, int(19052911)).Encode())
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
 * 心跳检测
 */
func (client *BiliClient) heartbeat() {
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
			heartbeatMsg := NewHeartbeatMessage(client.roomID, client.ownerUID)
			fmt.Println("heart")
			_, err := client.conn.Write(heartbeatMsg.Encode())
			if err != nil {
				log.Error("heartbeat failed, " + err.Error())
			}
		}
	}
}

func (client *BiliClient) chatMsg() {
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
				close(client.closeFlag)
				break loop
			}

			fmt.Println(string(b))
			// var jm map[string]interface{}
			// if err := json.Unmarshal(b, &jm); err != nil {
			// 	return nil, err
			// }
			// msgJSON, err = simplejson.NewJson(b)
			// cmd, err := msgJSON.Get("cmd").String()
			// if err == nil && cmd == danmuMSG {
			// 	fmt.Println(string(b))
			// }

			// analize message
		}
	}
}

/**
 * Read data from connection and process
 */
func (client *BiliClient) ReceiveMsg() ([]byte, int, error) {
	buf := make([]byte, 512)
	if _, err := io.ReadFull(client.conn, buf[:headerLENGTH]); err != nil {
		return buf, -1, err
	}

	// header
	// 4byte for packet length
	pl := binary.BigEndian.Uint32(buf[:4])

	// ignore buf[4:6] and buf[6:8]
	code := int(binary.BigEndian.Uint32(buf[8:12]))
	// ignore buf[12:16]

	// body content length
	cl := pl - headerLENGTH

	if cl > 512 {
		// expand buffer
		buf = make([]byte, cl)
	}
	if _, err := io.ReadFull(client.conn, buf[:cl]); err != nil {
		return buf, code, err
	}
	return buf[:cl], code, nil
}
