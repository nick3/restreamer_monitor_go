package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/sirupsen/logrus"
)

// BilibiliService provides access to Bilibili live streaming API
type BilibiliService struct {
	RoomId string
	Client *resty.Client
	logger *logrus.Entry
}

const (
	baseURL = "https://api.live.bilibili.com"
	userAgent = "Mozilla/5.0 (iPod; CPU iPhone OS 14_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/87.0.4280.163 Mobile/15E148 Safari/604.1"
	roomInitURL = "room/v1/Room/room_init"
	roomPlayInfoURL = "xlive/web-room/v2/index/getRoomPlayInfo"
	playURL = "room/v1/Room/playUrl"
	maxRetryCount = 3
	retryWaitTime = 5 * time.Second
	requestTimeout = 30 * time.Second
)

// validateRoomID validates the room ID format
func validateRoomID(roomID string) error {
	if roomID == "" {
		return fmt.Errorf("room ID cannot be empty")
	}
	
	// Check if room ID contains only digits
	matched, err := regexp.MatchString(`^\d+$`, roomID)
	if err != nil {
		return fmt.Errorf("failed to validate room ID: %w", err)
	}
	if !matched {
		return fmt.Errorf("room ID must contain only digits")
	}
	
	// Check reasonable length limits
	if len(roomID) > 20 {
		return fmt.Errorf("room ID is too long")
	}
	
	return nil
}

// NewBilibiliService creates a new BilibiliService instance with proper validation
func NewBilibiliService(roomId string) (*BilibiliService, error) {
	if err := validateRoomID(roomId); err != nil {
		return nil, fmt.Errorf("invalid room ID: %w", err)
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("User-Agent", userAgent).
		SetTimeout(requestTimeout).
		SetRetryCount(maxRetryCount).
		SetRetryWaitTime(retryWaitTime)

	return &BilibiliService{
		Client: client,
		RoomId: roomId,
		logger: logger.GetLogger(map[string]interface{}{
			"component": "service",
			"platform":  "bilibili",
			"room_id":   roomId,
		}),
	}, nil
}

// GetBilibiliRealRoomId retrieves the real room ID from Bilibili API
func (b *BilibiliService) GetBilibiliRealRoomId() (string, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
			"id": b.RoomId,
		}).
		Get(roomInitURL)

	if err != nil {
		return "", fmt.Errorf("failed to get room info: %w", err)
	}

	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			RoomId int `json:"room_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if data.Code != 0 {
		return "", fmt.Errorf("API error (code %d): %s", data.Code, data.Msg)
	}

	if data.Msg == "直播间不存在" {
		b.logger.WithField("room_id", b.RoomId).Warn("Room does not exist")
		return "", fmt.Errorf("room %s does not exist", b.RoomId)
	}
	
	return strconv.Itoa(data.Data.RoomId), nil
}

// GetBilibiliLiveStatus retrieves the live status of the room
func (b *BilibiliService) GetBilibiliLiveStatus() (bool, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
			"id": b.RoomId,
		}).
		Get(roomInitURL)

	if err != nil {
		return false, fmt.Errorf("failed to get live status: %w", err)
	}

	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			LiveStatus int `json:"live_status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.Code != 0 {
		return false, fmt.Errorf("API error (code %d): %s", data.Code, data.Msg)
	}

	if data.Msg == "直播间不存在" {
		b.logger.WithField("room_id", b.RoomId).Warn("Room does not exist")
		return false, fmt.Errorf("room %s does not exist", b.RoomId)
	}
	
	isLive := data.Data.LiveStatus == 1
	status := "offline"
	if isLive {
		status = "live"
	}
	b.logger.WithFields(logrus.Fields{
		"room_id": b.RoomId,
		"status":  status,
	}).Info("Room status check completed")
	return isLive, nil
}


// GetRoomBaseInfo retrieves the base information of the room and anchor
// This matches the bilicaptain library implementation
func (b *BilibiliService) GetRoomBaseInfo() (*struct {
	UID   string `json:"uid"`
	UName string `json:"uname"`
}, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
			"room_ids": b.RoomId,
			"req_biz":  "space",
		}).
		Get("xlive/web-room/v1/index/getRoomBaseInfo")

	if err != nil {
		return nil, fmt.Errorf("failed to get room base info: %w", err)
	}

	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ByRoomIds map[string]struct {
				UID       int    `json:"uid"`
				Uname     string `json:"uname"`
				Title     string `json:"title"`
				Cover     string `json:"cover"`
				LiveStart int64  `json:"live_start"`
			} `json:"by_room_ids"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", data.Code, data.Msg)
	}

	// Get the room data from by_room_ids map using room_id as key
	// Note: The key might be the real room ID, not the short ID
	roomInfo, exists := data.Data.ByRoomIds[b.RoomId]
	if !exists {
		// Try each key in the map to find a match
		for key, roomData := range data.Data.ByRoomIds {
			b.logger.WithFields(logrus.Fields{
				"found_id":    key,
				"searching_id": b.RoomId,
			}).Debug("Found room in response")
			roomInfo = roomData
			exists = true
			break
		}
	}

	if !exists || roomInfo.UID == 0 {
		return nil, fmt.Errorf("room %s not found in API response", b.RoomId)
	}

	return &struct {
		UID   string `json:"uid"`
		UName string `json:"uname"`
	}{
		UID:   strconv.Itoa(roomInfo.UID),
		UName: roomInfo.Uname,
	}, nil
}

// GetRoomInfo retrieves detailed room information
func (b *BilibiliService) GetRoomInfo() (*struct {
	Title     string    `json:"title"`
	UserCover string    `json:"user_cover"`
	Keyframe  string    `json:"keyframe"`
	LiveStart time.Time `json:"live_start"`
}, error) {
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
			"room_id": b.RoomId,
		}).
		Get("room/v1/Room/get_info")

	if err != nil {
		return nil, fmt.Errorf("failed to get room info: %w", err)
	}

	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Title     string `json:"title"`
			UserCover string `json:"user_cover"`
			Keyframe  string `json:"keyframe"`
			LiveTime  string `json:"live_time"` // Format: "YYYY-MM-DD HH:mm:ss"
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if data.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", data.Code, data.Msg)
	}

	// Parse live start time
	var liveStart time.Time
	if data.Data.LiveTime != "" && data.Data.LiveTime != "0000-00-00 00:00:00" {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", data.Data.LiveTime)
		if err != nil {
			b.logger.WithError(err).WithField("live_time", data.Data.LiveTime).Warn("Failed to parse live time")
		} else {
			liveStart = parsedTime
		}
	}

	return &struct {
		Title     string    `json:"title"`
		UserCover string    `json:"user_cover"`
		Keyframe  string    `json:"keyframe"`
		LiveStart time.Time `json:"live_start"`
	}{
		Title:     data.Data.Title,
		UserCover: data.Data.UserCover,
		Keyframe:  data.Data.Keyframe,
		LiveStart: liveStart,
	}, nil
}

// GetBilibiliLiveRealURL retrieves the real live stream URLs
func (b *BilibiliService) GetBilibiliLiveRealURL(realRoomId string) ([]string, error) {
	if err := validateRoomID(realRoomId); err != nil {
		return nil, fmt.Errorf("invalid real room ID: %w", err)
	}

	// processURL converts FLV URLs to M3U8 format
	processURL := func(urlStr string) string {
		u, err := url.Parse(urlStr)
		if err != nil {
			b.logger.WithError(err).WithField("url", urlStr).Warn("Failed to parse URL")
			return urlStr
		}
		pathParts := strings.Split(u.Path, "/")
		if len(pathParts) == 0 {
			return urlStr
		}
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

	// Try playUrl API first
	resp, err := b.Client.R().
		SetQueryParams(map[string]string{
			"cid":      realRoomId,
			"qn":       "10000",
			"platform": "web",
		}).
		Get(playURL)

	if err != nil {
		return nil, fmt.Errorf("failed to get play URL: %w", err)
	}

	var playData struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Durl []struct {
				URL string `json:"url"`
			} `json:"durl"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(resp.Body(), &playData); err != nil {
		return nil, fmt.Errorf("failed to parse play URL response: %w", err)
	}

	if playData.Code != 0 {
		b.logger.WithFields(logrus.Fields{
			"code": playData.Code,
			"msg":  playData.Msg,
		}).Warn("Play URL API returned error")
	} else if len(playData.Data.Durl) > 0 {
		url1 := processURL(playData.Data.Durl[0].URL)
		url2 := playData.Data.Durl[0].URL
		return []string{url1, url2}, nil
	}

	// Fallback to room play info API
	resp, err = b.Client.R().
		SetQueryParams(map[string]string{
			"room_id":    realRoomId,
			"no_playurl": "0",
			"mask":       "0",
			"qn":         "10000",
			"platform":   "web",
			"protocol":   "0,1",
			"format":     "0,1,2",
			"codec":      "0,1",
		}).
		Get(roomPlayInfoURL)

	if err != nil {
		return nil, fmt.Errorf("failed to get room play info: %w", err)
	}

	var roomData struct {
		Code int `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PlayUrlInfo struct {
				PlayUrl struct {
					Durl []struct {
						URL string `json:"url"`
					} `json:"durl"`
				} `json:"playurl"`
			} `json:"playurl_info"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body(), &roomData); err != nil {
		b.logger.WithFields(logrus.Fields{
			"error":           err,
			"response_length": len(resp.Body()),
		}).Error("Failed to parse room play info response")
		return nil, fmt.Errorf("failed to parse room play info response: %w", err)
	}

	if roomData.Code != 0 {
		return nil, fmt.Errorf("room play info API error (code %d): %s", roomData.Code, roomData.Msg)
	}

	if len(roomData.Data.PlayUrlInfo.PlayUrl.Durl) > 0 {
		urls := make([]string, 0, len(roomData.Data.PlayUrlInfo.PlayUrl.Durl))
		for _, durl := range roomData.Data.PlayUrlInfo.PlayUrl.Durl {
			urls = append(urls, processURL(durl.URL))
		}
		return urls, nil
	}

	return nil, fmt.Errorf("no live stream URLs found for room %s", realRoomId)
}