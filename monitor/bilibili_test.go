package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBilibiliStreamSource(t *testing.T) {
	t.Run("valid room ID", func(t *testing.T) {
		source, err := NewBilibiliStreamSource("123")
		require.NoError(t, err)
		assert.NotNil(t, source)
		assert.Equal(t, "123", source.roomInfo.RoomID)
		assert.Equal(t, "bilibili", source.roomInfo.Platform)
	})

	t.Run("invalid room ID", func(t *testing.T) {
		source, err := NewBilibiliStreamSource("invalid")
		assert.Error(t, err)
		assert.Nil(t, source)
	})
}

func TestBilibiliStreamSource_GetStatus(t *testing.T) {
	source, err := NewBilibiliStreamSource("76")
	require.NoError(t, err)

	// First call should work (may return false if room is offline)
	status := source.GetStatus()
	assert.IsType(t, false, status)

	// Second call should also work
	status2 := source.GetStatus()
	assert.IsType(t, false, status2)
}

func TestBilibiliStreamSource_GetRoomInfo(t *testing.T) {
	source, err := NewBilibiliStreamSource("76")
	require.NoError(t, err)

	roomInfo := source.GetRoomInfo()
	assert.Equal(t, "76", roomInfo.RoomID)
	assert.Equal(t, "bilibili", roomInfo.Platform)
	// RealRoomID should be populated after the call
	assert.NotEmpty(t, roomInfo.RealRoomID)
}

func TestBilibiliStreamSource_GetPlayURL(t *testing.T) {
	source, err := NewBilibiliStreamSource("76")
	require.NoError(t, err)

	// This may return empty string if room is offline
	playURL := source.GetPlayURL()
	assert.IsType(t, "", playURL)
}

func TestBilibiliStreamSource_MessageListener(t *testing.T) {
	source, err := NewBilibiliStreamSource("76")
	require.NoError(t, err)

	// These should not panic
	assert.NotPanics(t, func() {
		source.StartMsgListener()
	})

	assert.NotPanics(t, func() {
		source.CloseMsgListener()
	})
}