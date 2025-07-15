package models

import "time"

// RoomInfo represents live room information
type RoomInfo struct {
	Platform    string    `json:"platform"`
	RoomID      string    `json:"room_id"`
	UID         string    `json:"uid"`
	UName       string    `json:"uname"`
	RealRoomID  string    `json:"real_room_id"`
	IsLive      bool      `json:"is_live"`
	UserCover   string    `json:"user_cover"`
	Keyframe    string    `json:"keyframe"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
}