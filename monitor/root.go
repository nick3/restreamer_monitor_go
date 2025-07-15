package monitor

import (
	"github.com/nick3/restreamer_monitor_go/models"
)

// StreamSource defines the interface for live stream sources
type StreamSource interface {
	GetStatus() bool
	GetRoomInfo() models.RoomInfo
	GetPlayURL() string
	StartMsgListener()
	CloseMsgListener()
}