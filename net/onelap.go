package net

import (
	"encoding/json"
	"fmt"

	"danmu/config"
)

type UserRow struct {
	Uid         int     `json:"uid"`
	UserName    string  `json:"user_name"`
	Email       string  `json:"email"`
	Mobile      string  `json:"mobile"`
	Sex         string  `json:"sex"`
	Age         int     `json:"age"`
	Level       int     `json:"level"`
	Vip         string  `json:"vip"`
	Pay         float32 `json:"pay"`
	OnlineTime  int     `json:"online_time"`
	OnlineCount int     `json:"online_count"`
	Location    string  `json:"location"`
	Fitness     Device  `json:"fitness"`
	Power       Device  `json:"power"`
	Speed       Device  `json:"speed"`
	Heart       Device  `json:"heart"`
	Cadence     Device  `json:"cadence"`
}

type Device struct {
	FactureId string `json:"factureId"`
	AntId     string `json:"antId"`
	Time      int    `json:"time"`
	LastRide  string `json:"last_ride"`
}

/**
 *
 * 获取用户详情接口
 *
 */
func GetUserInfo(uid int) ([]UserRow, error) {
	url, ok := config.Conf.Get("api", "api_user_info")
	url = fmt.Sprintf(url, uid)
	var rows []UserRow
	if ok {
		data := Get(url)
		err := json.Unmarshal(data, &rows)
		if err != nil {
			return rows, err
		}
	}
	return rows, nil
}

/**
 * 获取指定数量用户ID
 *
 */
func GetUids() ([]int, error) {
	url, ok := config.Conf.Get("api", "api_uids")
	url = fmt.Sprintf(url)
	fmt.Println(url)
	var uids []int
	if ok {
		data := Get(url)
		err := json.Unmarshal(data, &uids)
		if err != nil {
			return uids, err
		}
	}
	return uids, nil
}

func init() {

}
