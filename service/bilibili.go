package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// BilibiliService is a service for Bilibili API.

const BASE_URL = "https://api.live.bilibili.com"
var rrURL = "room/v1/Room/room_init"
var ruURL = "xlive/web-room/v2/index/getRoomPlayInfo"
var playURL = "room/v1/Room/playUrl"

type BilibiliService struct {
	RoomId string
	Client *resty.Client
}

func NewBilibiliService(roomId string) *BilibiliService {
	client := resty.New().
		SetBaseURL(BASE_URL).
		SetHeader("User-Agent", "Mozilla/5.0 (iPod; CPU iPhone OS 14_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/87.0.4280.163 Mobile/15E148 Safari/604.1").
		SetTimeout(30 * time.Second).
        SetRetryCount(3).                // 添加重试次数
        SetRetryWaitTime(5 * time.Second) // 重试等待时间
    return &BilibiliService{
        Client: client,
        RoomId: roomId,
    }
}

// 获取真实的房间号
func (b *BilibiliService) GetBilibiliRealRoomId() (string, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
          "id": b.RoomId,
      	}).
		Get(rrURL)

    if err != nil {
        return "", err
    }

    var data struct {
        Code int `json:"code"`
        Msg  string `json:"msg"`
        Data struct {
            RoomId int `json:"room_id"`
        } `json:"data"`
    }

	err = json.Unmarshal(resp.Body(), &data)
    if err != nil {
        return "", err
    }

    if data.Msg == "直播间不存在" {
        log.Printf("%s 直播间不存在", b.RoomId)
        return "", fmt.Errorf("%s 直播间不存在", b.RoomId)
    }
    return strconv.Itoa(data.Data.RoomId), nil
}

// 获取直播状态
func (b *BilibiliService) GetBilibiliLiveStatus() (bool, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
          "id": b.RoomId,
      	}).
		Get(rrURL)

    if err != nil {
        return false, err
    }

    var data struct {
        Code int `json:"code"`
        Msg  string `json:"msg"`
        Data struct {
            LiveStatus int `json:"live_status"`
        } `json:"data"`
    }

    err = json.Unmarshal(resp.Body(), &data)
    if err != nil {
        return false, err
    }

    if data.Msg == "直播间不存在" {
        log.Printf("%s 直播间不存在", b.RoomId)
        return false, fmt.Errorf("%s 直播间不存在", b.RoomId)
    }
    isLive := data.Data.LiveStatus == 1
    log.Printf("%s %s", b.RoomId, map[bool]string{true: "正在直播", false: "未直播"}[isLive])
    return isLive, nil
}


// 获取真实的直播流 URL
func (b *BilibiliService) GetBilibiliLiveRealURL(realRoomId string) ([]string, error) {
    // 实现逻辑与 TypeScript 代码类似
    // 处理 URL 的函数
    processURL := func(urlStr string) string {
        u, err := url.Parse(urlStr)
        if err != nil {
            return urlStr
        }
        pathParts := strings.Split(u.Path, "/")
        filename := pathParts[len(pathParts)-1]
        if strings.HasSuffix(filename, ".flv") {
            filename = strings.TrimSuffix(filename, ".flv")
            filename += "/index.m3u8"
            pathParts[len(pathParts)-1] = filename
            u.Path = strings.Join(pathParts, "/")
            u.RawQuery = ""
            return u.String()
        }
        return urlStr
    }

    // 首先请求 playUrl 接口
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
          "cid": realRoomId,
		  "qn": "10000",
		  "platform": "web",
      	}).
		Get(playURL)

    if err != nil {
        return nil, err
    }

    var data struct {
        Data struct {
            Durl []struct {
                URL string `json:"url"`
            } `json:"durl"`
        } `json:"data"`
    }
    err = json.Unmarshal(resp.Body(), &data)
    if err != nil {
        return nil, err
    }

    if len(data.Data.Durl) > 0 {
        url1 := processURL(data.Data.Durl[0].URL)
        url2 := data.Data.Durl[0].URL
        return []string{url1, url2}, nil
    }

    // 如果没有数据，尝试请求 ruUrl 接口
    resp, err = b.Client.R().
		SetQueryParams(map[string]string{
			"room_id": realRoomId,
			"no_playurl": "0",
			"mask": "0",
			"qn": "10000",
			"platform": "web",
			"protocol": "0,1",
			"format": "0,1,2",
			"codec": "0,1",
		}).
		Get(ruURL)

	if err != nil {
		return nil, err
	}

	// 打印出 resp.Body()，看看返回的数据结构
	log.Println(string(resp.Body()))

    return nil, fmt.Errorf("未能获取到直播流 URL")
}