package telegram

import (
	"fmt"
	"testing"
	"time"

	"github.com/nick3/restreamer_monitor_go/models"
)

func TestFormatLiveStartNotification(t *testing.T) {
	tests := []struct {
		name     string
		roomInfo models.RoomInfo
		wantText bool
		wantPic  bool
	}{
		{
			"full room info",
			models.RoomInfo{
				Platform:    "bilibili",
				RoomID:      "123",
				RealRoomID:  "456",
				UID:         "789",
				UName:       "Test主播",
				Title:       "测试直播间",
				UserCover:   "http://example.com/cover.jpg",
				Keyframe:    "http://example.com/keyframe.jpg",
				StartTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				IsLive:      true,
			},
			true,
			true,
		},
		{
			"minimal room info",
			models.RoomInfo{
				Platform:  "bilibili",
				RoomID:    "123",
				UName:     "Test主播",
				IsLive:    true,
			},
			true,
			false, // No image URL
		},
		{
			"with keyframe fallback",
			models.RoomInfo{
				Platform:   "bilibili",
				RoomID:     "123",
				RealRoomID: "456",
				UName:      "Test主播",
				Title:      "测试直播间",
				Keyframe:   "http://example.com/keyframe.jpg",
				StartTime:  time.Now(),
				IsLive:     true,
			},
			true,
			true, // Should use keyframe as fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, photoURL := FormatLiveStartNotification(tt.roomInfo)

			if tt.wantText {
				if message == "" {
					t.Error("Expected non-empty message")
				}

				// Check for required elements in message
				if tt.roomInfo.UName != "" && !contains(message, tt.roomInfo.UName) {
					t.Errorf("Message should contain UName %s", tt.roomInfo.UName)
				}

				if tt.roomInfo.Title != "" && !contains(message, tt.roomInfo.Title) {
					t.Errorf("Message should contain Title %s", tt.roomInfo.Title)
				}

				t.Logf("Formatted message:\n%s", message)
			}

			if tt.wantPic {
				if photoURL == "" {
					t.Error("Expected non-empty photo URL")
				}
				t.Logf("Photo URL: %s", photoURL)
			} else {
				if photoURL != "" {
					t.Errorf("Expected empty photo URL, got %s", photoURL)
				}
			}
		})
	}
}

func TestFormatLiveEndNotification(t *testing.T) {
	tests := []struct {
		name     string
		roomInfo models.RoomInfo
		wantText bool
	}{
		{
			"full room info",
			models.RoomInfo{
				Platform:   "bilibili",
				RoomID:     "123",
				RealRoomID: "456",
				UID:        "789",
				UName:      "Test主播",
				IsLive:     false,
			},
			true,
		},
		{
			"minimal room info",
			models.RoomInfo{
				Platform: "bilibili",
				RoomID:   "123",
				UName:    "Test主播",
				IsLive:   false,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := FormatLiveEndNotification(tt.roomInfo)

			if tt.wantText {
				if message == "" {
					t.Error("Expected non-empty message")
				}

				// Check for required elements in message
				if tt.roomInfo.UName != "" && !contains(message, tt.roomInfo.UName) {
					t.Errorf("Message should contain UName %s", tt.roomInfo.UName)
				}

				if tt.roomInfo.UID != "" && !contains(message, tt.roomInfo.UID) {
					t.Errorf("Message should contain UID %s", tt.roomInfo.UID)
				}

				t.Logf("Formatted message:\n%s", message)
			}
		})
	}
}

func TestFormatStatusNotification(t *testing.T) {
	status := "System running"
	details := map[string]interface{}{
		"room_count": 5,
		"uptime":     "2h 30m",
		"memory":     "256 MB",
	}

	message := FormatStatusNotification(status, details)

	if message == "" {
		t.Error("Expected non-empty message")
	}

	if !contains(message, status) {
		t.Errorf("Message should contain status %s", status)
	}

	for key, value := range details {
		if !contains(message, key) {
			t.Errorf("Message should contain detail key %s", key)
		}
		strValue := fmt.Sprintf("%v", value)
		if !contains(message, strValue) {
			t.Errorf("Message should contain detail value %s", strValue)
		}
	}

	t.Logf("Formatted message:\n%s", message)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr)))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
