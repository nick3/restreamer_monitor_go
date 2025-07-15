package monitor

import (
	"log"
	"time"

	"github.com/nick3/restreamer_monitor_go/models"
	"github.com/nick3/restreamer_monitor_go/service"
)

// BilibiliStreamSource implements StreamSource interface for Bilibili platform
type BilibiliStreamSource struct {
	service    *service.BilibiliService
	roomInfo   models.RoomInfo
	lastStatus bool
}

// NewBilibiliStreamSource creates a new Bilibili stream source
func NewBilibiliStreamSource(roomID string) (*BilibiliStreamSource, error) {
	svc, err := service.NewBilibiliService(roomID)
	if err != nil {
		return nil, err
	}

	return &BilibiliStreamSource{
		service: svc,
		roomInfo: models.RoomInfo{
			Platform: "bilibili",
			RoomID:   roomID,
		},
	}, nil
}

// GetStatus returns the current live status
func (b *BilibiliStreamSource) GetStatus() bool {
	status, err := b.service.GetBilibiliLiveStatus()
	if err != nil {
		log.Printf("Failed to get live status: %v", err)
		return false
	}
	
	// Update room info if status changed
	if status != b.lastStatus {
		b.roomInfo.IsLive = status
		if status {
			b.roomInfo.StartTime = time.Now()
		}
		b.lastStatus = status
	}
	
	return status
}

// GetRoomInfo returns the room information
func (b *BilibiliStreamSource) GetRoomInfo() models.RoomInfo {
	// Update real room ID if not set
	if b.roomInfo.RealRoomID == "" {
		realRoomID, err := b.service.GetBilibiliRealRoomId()
		if err != nil {
			log.Printf("Failed to get real room ID: %v", err)
		} else {
			b.roomInfo.RealRoomID = realRoomID
		}
	}
	
	return b.roomInfo
}

// GetPlayURL returns the live stream URL
func (b *BilibiliStreamSource) GetPlayURL() string {
	realRoomID := b.roomInfo.RealRoomID
	if realRoomID == "" {
		var err error
		realRoomID, err = b.service.GetBilibiliRealRoomId()
		if err != nil {
			log.Printf("Failed to get real room ID: %v", err)
			return ""
		}
		b.roomInfo.RealRoomID = realRoomID
	}
	
	urls, err := b.service.GetBilibiliLiveRealURL(realRoomID)
	if err != nil {
		log.Printf("Failed to get live URLs: %v", err)
		return ""
	}
	
	if len(urls) > 0 {
		return urls[0] // Return the first URL (usually M3U8 format)
	}
	
	return ""
}

// StartMsgListener starts listening for live messages (placeholder)
func (b *BilibiliStreamSource) StartMsgListener() {
	log.Printf("Starting message listener for room %s", b.roomInfo.RoomID)
	// TODO: Implement WebSocket connection for live messages
}

// CloseMsgListener closes the message listener (placeholder)
func (b *BilibiliStreamSource) CloseMsgListener() {
	log.Printf("Closing message listener for room %s", b.roomInfo.RoomID)
	// TODO: Implement WebSocket connection cleanup
}