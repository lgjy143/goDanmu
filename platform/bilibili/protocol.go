package bilibili

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	headerLENGTH = 16 // in bytes
	deviceTYPE   = 1
	device       = 1
)

const (
	// cmd types
	danmuMSG     = "DANMU_MSG"
	danmuGIFT    = "DANMU_GIFT"
	danmuWelcome = "WELCOME"
	DANMU_MSG    = "DANMU_MSG"
	// 停播
	LIVE_OFF = 0
	// 直播
	LIVE_ON = 1
	// 轮播
	LiVE_ROTATE = 2
)

type Message struct {
	body     []byte
	bodyType int32
}

func NewHandshakeMessage(roomid, uid int) *Message {

	data := fmt.Sprintf(`{"roomid":%d,"uid":%d}`, roomid, uid)
	message := &Message{
		body:     []byte(data),
		bodyType: 7,
	}
	return message

}

func NewHeartbeatMessage() *Message {

	data := ""
	message := &Message{
		body:     []byte(data),
		bodyType: 2,
	}
	return message

}

func NewMessage(b []byte, btype int) *Message {
	return &Message{
		body:     b,
		bodyType: int32(btype),
	}

}

func (msg *Message) Encode() []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, int32(len(msg.body)+headerLENGTH)) // write package length
	binary.Write(buffer, binary.BigEndian, int16(headerLENGTH))               // header length
	binary.Write(buffer, binary.BigEndian, int16(deviceTYPE))
	binary.Write(buffer, binary.BigEndian, int32(msg.bodyType))
	binary.Write(buffer, binary.BigEndian, int32(device))
	binary.Write(buffer, binary.BigEndian, msg.body)
	return buffer.Bytes()
}

func (msg *Message) Decode() *Message {
	// TODO
	return msg
}

// func (msg *Message) GetCmd() string {
// 	jc, err := rrconfig.LoadJsonConfigFromBytes(msg.body)
// 	if err != nil {
// 		log.Error(err)
// 		return "INVALID"
// 	}
// 	cmd, err := jc.GetString("cmd")
// 	if err != nil {
// 		log.Error(err)
// 		return "ERROR"
// 	}
// 	return cmd

// }

func (msg *Message) Bytes() []byte {
	return msg.body
}
