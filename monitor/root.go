package monitor

import (
	"github.com/nick3/restreamer_monitor_go/models"
)

type StreamSource interface {
    getStatus() bool
	getRoomInfo() models.RoomInfo
	getPlayURL() string
	startMsgListerner()
	closeMsgListerner()
}