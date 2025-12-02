package monitor

import (
	"fmt"
	"time"

	"github.com/nick3/restreamer_monitor_go/logger"
	"github.com/nick3/restreamer_monitor_go/models"
	"github.com/nick3/restreamer_monitor_go/service"
	"github.com/sirupsen/logrus"
)

// BilibiliStreamSource implements StreamSource interface for Bilibili platform
type BilibiliStreamSource struct {
	service    *service.BilibiliService
	roomInfo   models.RoomInfo
	lastStatus bool
	logger     *logrus.Entry
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
		logger: logger.GetLogger(map[string]interface{}{
			"component": "monitor",
			"platform":  "bilibili",
			"room_id":   roomID,
		}),
	}, nil
}

// GetStatus returns the current live status
func (b *BilibiliStreamSource) GetStatus() bool {
	status, err := b.service.GetBilibiliLiveStatus()
	if err != nil {
		b.logger.WithError(err).Error("Failed to get live status")
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
			b.logger.WithError(err).Error("Failed to get real room ID")
		} else {
			b.roomInfo.RealRoomID = realRoomID
		}
	}

	// Fetch detailed room information if not already populated
	// Fallback approach: try multiple methods to get anchor info
	if b.roomInfo.UID == "" || b.roomInfo.UName == "" {
		// Try primary method: GetRoomBaseInfo (rich info, but may be rate-limited)
		if baseInfo, err := b.service.GetRoomBaseInfo(); err == nil {
			b.roomInfo.UID = baseInfo.UID
			b.roomInfo.UName = baseInfo.UName
		} else {
			b.logger.WithError(err).Error("Failed to get room base info")

			// Fallback: use default values to ensure notifications still work
			if b.roomInfo.UName == "" {
				b.roomInfo.UName = fmt.Sprintf("主播%s", b.roomInfo.RoomID)
				b.logger.WithField("anchor_name", b.roomInfo.UName).Warn("Using default anchor name")
			}
		}
	}

	// Get room title and cover (this API is more stable)
	if b.roomInfo.Title == "" || b.roomInfo.UserCover == "" {
		if roomInfo, err := b.service.GetRoomInfo(); err == nil {
			b.roomInfo.Title = roomInfo.Title
			b.roomInfo.UserCover = roomInfo.UserCover
			b.roomInfo.Keyframe = roomInfo.Keyframe
			if b.roomInfo.StartTime.IsZero() {
				b.roomInfo.StartTime = roomInfo.LiveStart
			}
		} else {
			b.logger.WithError(err).Error("Failed to get room info")
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
			b.logger.WithError(err).Error("Failed to get real room ID")
			return ""
		}
		b.roomInfo.RealRoomID = realRoomID
	}
	
	urls, err := b.service.GetBilibiliLiveRealURL(realRoomID)
	if err != nil {
		b.logger.WithError(err).Error("Failed to get live URLs")
		return ""
	}

	if len(urls) > 0 {
		return urls[0] // Return the first URL (usually M3U8 format)
	}

	return ""
}

// StartMsgListener starts listening for live messages (placeholder)
func (b *BilibiliStreamSource) StartMsgListener() {
	b.logger.WithField("room_id", b.roomInfo.RoomID).Info("Starting message listener")
	// TODO: Implement WebSocket connection for live messages
}

// CloseMsgListener closes the message listener (placeholder)
func (b *BilibiliStreamSource) CloseMsgListener() {
	b.logger.WithField("room_id", b.roomInfo.RoomID).Info("Closing message listener")
	// TODO: Implement WebSocket connection cleanup
}