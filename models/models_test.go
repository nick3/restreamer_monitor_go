package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRoomInfo(t *testing.T) {
	t.Run("create room info", func(t *testing.T) {
		now := time.Now()
		roomInfo := RoomInfo{
			Platform:   "bilibili",
			RoomID:     "123",
			UID:        "456",
			UName:      "TestUser",
			RealRoomID: "789",
			IsLive:     true,
			UserCover:  "http://example.com/cover.jpg",
			Keyframe:   "http://example.com/keyframe.jpg",
			Title:      "Test Stream",
			StartTime:  now,
		}

		assert.Equal(t, "bilibili", roomInfo.Platform)
		assert.Equal(t, "123", roomInfo.RoomID)
		assert.Equal(t, "456", roomInfo.UID)
		assert.Equal(t, "TestUser", roomInfo.UName)
		assert.Equal(t, "789", roomInfo.RealRoomID)
		assert.True(t, roomInfo.IsLive)
		assert.Equal(t, "http://example.com/cover.jpg", roomInfo.UserCover)
		assert.Equal(t, "http://example.com/keyframe.jpg", roomInfo.Keyframe)
		assert.Equal(t, "Test Stream", roomInfo.Title)
		assert.Equal(t, now, roomInfo.StartTime)
	})

	t.Run("zero value", func(t *testing.T) {
		var roomInfo RoomInfo
		assert.Empty(t, roomInfo.Platform)
		assert.Empty(t, roomInfo.RoomID)
		assert.False(t, roomInfo.IsLive)
		assert.True(t, roomInfo.StartTime.IsZero())
	})
}