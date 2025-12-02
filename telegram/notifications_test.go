package telegram

import (
	"fmt"
	"strings"
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
				if tt.roomInfo.UName != "" && !strings.Contains(message, tt.roomInfo.UName) {
					t.Errorf("Message should contain UName %s", tt.roomInfo.UName)
				}

				if tt.roomInfo.Title != "" && !strings.Contains(message, tt.roomInfo.Title) {
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
				if tt.roomInfo.UName != "" && !strings.Contains(message, tt.roomInfo.UName) {
					t.Errorf("Message should contain UName %s", tt.roomInfo.UName)
				}

				if tt.roomInfo.UID != "" && !strings.Contains(message, tt.roomInfo.UID) {
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
		"roomcount": 5, // Use keys without special chars for basic test
		"uptime":    "2h 30m",
		"memory":    "256 MB",
	}

	message := FormatStatusNotification(status, details)

	if message == "" {
		t.Error("Expected non-empty message")
	}

	if !strings.Contains(message, status) {
		t.Errorf("Message should contain status %s", status)
	}

	for key, value := range details {
		// Key might be escaped, so check for the escaped version too
		escapedKey := escapeMarkdown(key)
		if !strings.Contains(message, key) && !strings.Contains(message, escapedKey) {
			t.Errorf("Message should contain detail key %s (or escaped: %s)", key, escapedKey)
		}
		strValue := fmt.Sprintf("%v", value)
		if !strings.Contains(message, strValue) {
			t.Errorf("Message should contain detail value %s", strValue)
		}
	}

	t.Logf("Formatted message:\n%s", message)
}

func TestFormatStatusNotification_WithSpecialChars(t *testing.T) {
	// Test that special MarkdownV2 characters are properly escaped
	status := "System [running] with *special* chars_here"
	details := map[string]interface{}{
		"test_key": "value",
		"key-2":    100,
	}

	message := FormatStatusNotification(status, details)

	if message == "" {
		t.Error("Expected non-empty message")
	}

	// Verify special characters are escaped
	if strings.Contains(message, "[running]") {
		t.Error("Square brackets should be escaped")
	}
	if strings.Contains(message, "*special*") && !strings.Contains(message, "\\*special\\*") {
		t.Error("Asterisks should be escaped in status text")
	}

	t.Logf("Formatted message with special chars:\n%s", message)
}
