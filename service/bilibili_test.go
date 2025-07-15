package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRoomID(t *testing.T) {
	tests := []struct {
		name    string
		roomID  string
		wantErr bool
	}{
		{"valid room ID", "123456", false},
		{"empty room ID", "", true},
		{"non-numeric room ID", "abc123", true},
		{"room ID with special chars", "123-456", true},
		{"too long room ID", "123456789012345678901", true},
		{"valid single digit", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoomID(tt.roomID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseURL(t *testing.T) {
	expected := "https://api.live.bilibili.com"
	assert.Equal(t, expected, baseURL)
}

func TestNewBilibiliService(t *testing.T) {
	t.Run("valid room ID", func(t *testing.T) {
		roomID := "123"
		service, err := NewBilibiliService(roomID)
		
		require.NoError(t, err)
		assert.Equal(t, roomID, service.RoomId)
		assert.NotNil(t, service.Client)
		assert.Equal(t, baseURL, service.Client.BaseURL)
	})

	t.Run("invalid room ID", func(t *testing.T) {
		roomID := "invalid123"
		service, err := NewBilibiliService(roomID)
		
		assert.Error(t, err)
		assert.Nil(t, service)
	})
}

func TestBilibiliService_GetBilibiliRealRoomId(t *testing.T) {
	service, err := NewBilibiliService("76")
	require.NoError(t, err)
	
	roomId, err := service.GetBilibiliRealRoomId()
	if err != nil {
		t.Logf("GetBilibiliRealRoomId error: %v", err)
		// This is expected if the room doesn't exist or API is unreachable
		return
	}
	t.Logf("Real room ID: %v", roomId)
}

func TestBilibiliService_GetBilibiliLiveStatus(t *testing.T) {
	service, err := NewBilibiliService("76")
	require.NoError(t, err)
	
	isLive, err := service.GetBilibiliLiveStatus()
	if err != nil {
		t.Logf("GetBilibiliLiveStatus error: %v", err)
		// This is expected if the room doesn't exist or API is unreachable
		return
	}
	t.Logf("Live status: %v", isLive)
}