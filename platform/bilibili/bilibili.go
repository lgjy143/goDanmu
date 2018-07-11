package bilibili

import (
	"danmu/utils"
	"danmu/utils/log"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/golang/glog"
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
	reURL := regexp.MustCompile(`.*\/(\d+)[?]*.*$`)
	roomArr := reURL.FindStringSubmatch(client.originURL)
	if len(roomArr) == 2 {
		roomInitHTML := utils.Get(fmt.Sprintf(roomInitAddr, roomArr[1]))

		roomJSON, err := simplejson.NewJson([]byte(roomInitHTML))

		if err != nil {
			glog.Fatal(err.Error())
		}

		if resOk, _ := roomJSON.Get("msg").String(); resOk == "ok" {
			roomData := roomJSON.Get("data")
			client.roomID, _ = roomData.Get("room_id").Int()
			client.ownerUID, _ = roomData.Get("uid").Int()
			client.liveStatus, _ = roomData.Get("live_status").Int()
			glog.Infof("Room Live Status %d", client.liveStatus)
		} else {
			glog.Fatal("Room Init Failed")
		}

		if client.liveStatus == LIVE_OFF {
			glog.Info("Room is Closed")
			return
		}

		glog.Info("Get Barrage Servers Info")
		dmIP, dmPort, err := getBarrageServer(client.roomID)
		if err != nil {
			glog.Fatal(err)
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
	conn, err := net.Dial("tcp", danmuServerStr)
	if err != nil {
		return err
	}
	glog.Infof("%s connected.", danmuServerStr)

	client.conn = conn
	randUid := rand.Intn(200000) + 100000
	// hand shake
	client.conn.Write(NewHandshakeMessage(client.roomID, randUid).Encode())
	glog.Info("Connect Handshake")

	client.wg.Add(2)
	// heart
	go client.heartbeat()
	// go client.reHandShake()
	// chatMsg
	go client.chatMsg()
	client.wg.Wait()
	return nil
}

/**
 * 心跳检测
 */
func (client *BiliClient) heartbeat() {
	defer client.wg.Done()
	tick := time.Tick(5 * time.Second)
loop:
	for {
		select {
		case _, ok := <-client.closeFlag:
			if !ok {
				break loop
			}
		case <-tick:
			heartbeatMsg := NewHeartbeatMessage()
			fmt.Println("heart")
			_, err := client.conn.Write(heartbeatMsg.Encode())
			if err != nil {
				log.Error("heartbeat failed, " + err.Error())
			}
		}
	}
}

func (client *BiliClient) reConnect() error {
	var danmuServerStr = client.serverIP + ":" + client.serverPort
	conn, err := net.Dial("tcp", danmuServerStr)
	if err != nil {
		return err
	}
	glog.Infof("%s reconnected.", danmuServerStr)

	client.conn = conn
	randUid := rand.Intn(200000) + 100000
	// hand shake
	client.conn.Write(NewHandshakeMessage(client.roomID, randUid).Encode())
	glog.Info("ReConnect Handshake")
	return nil
}

func (client *BiliClient) reHandShake() {
	defer client.wg.Done()
	tick := time.Tick(50 * time.Second)
loop:
	for {
		select {
		case _, ok := <-client.closeFlag:
			if !ok {
				break loop
			}
		case <-tick:
			randUID := rand.Intn(200000) + 100000
			// hand shake
			fmt.Println("re handshake")
			_, err := client.conn.Write(NewHandshakeMessage(client.roomID, randUID).Encode())
			if err != nil {
				log.Error("rehandshake failed, " + err.Error())
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
				glog.Error(err, code)
				if code == -1 {
					client.reConnect()
				} else {
					close(client.closeFlag)
					break loop
				}
			}

			switch code {
			case 3:
				glog.Info("heartbeat ok")
				continue
			case 8:
				glog.Info("handshake ok")
				continue
			case 5:
				jsonStr := string(b)
				// log.Info(jsonStr)
				reCMD := regexp.MustCompile(`.*"cmd":"(\w+)"`)
				cmdArr := reCMD.FindStringSubmatch(jsonStr)
				if len(cmdArr) == 2 {
					fmt.Println(cmdArr[1])
					glog.Info(cmdArr[1])
				}
				reContent := regexp.MustCompile(`\],"(.*)",\[`)
				contentArr := reContent.FindStringSubmatch(jsonStr)
				if len(contentArr) == 2 {
					fmt.Println(contentArr[1])
					glog.Info(contentArr[1])
				}
				continue
			}

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
