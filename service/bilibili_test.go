package service

import (
	"testing"
)

func TestBaseURL(t *testing.T) {
	expected := "https://api.live.bilibili.com"
	if BASE_URL != expected {
		t.Errorf("BASE_URL = %s; want %s", BASE_URL, expected)
	}
}

func TestNewBilibiliService(t *testing.T) {
	roomID := "123"
	service := NewBilibiliService(roomID)

	if service.RoomId != roomID {
		t.Errorf("RoomId = %s; want %s", service.RoomId, roomID)
	}

	if service.Client == nil {
		t.Error("Client should not be nil")
	}

	if service.Client.BaseURL != BASE_URL {
		t.Errorf("Client BaseURL = %s; want %s", service.Client.BaseURL, BASE_URL)
	}
}

func TestBilibiliService_GetBilibiliRealRoomId(t *testing.T) {
    service := NewBilibiliService("76")
    roomId, err := service.GetBilibiliRealRoomId()
    if err != nil {
        t.Errorf("GetLiveURL error: %v", err)
    }
    t.Logf("roomId: %v", roomId)
}