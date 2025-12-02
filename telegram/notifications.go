package telegram

import (
	"fmt"
	"strings"
	"time"

	"github.com/nick3/restreamer_monitor_go/models"
)

// escapeMarkdown escapes special characters for Telegram MarkdownV2
// Reference: https://core.telegram.org/bots/api#markdownv2-style
func escapeMarkdown(text string) string {
	if text == "" {
		return ""
	}

	// Characters that need to be escaped in MarkdownV2
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

	escaped := text
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}

	return escaped
}

// FormatLiveStartNotification formats a notification for when a live stream starts
// Returns the text message and photo URL
func FormatLiveStartNotification(roomInfo models.RoomInfo) (string, string) {
	// Use real room ID for links if available, otherwise use configured room ID
	roomID := roomInfo.RealRoomID
	if roomID == "" {
		roomID = roomInfo.RoomID
	}

	// Format the live stream URL
	liveURL := fmt.Sprintf("https://live.bilibili.com/%s", roomID)

	// Format the start time
	timeStr := ""
	if !roomInfo.StartTime.IsZero() {
		timeStr = roomInfo.StartTime.Format("2006-01-02 15:04:05")
	} else {
		timeStr = time.Now().Format("2006-01-02 15:04:05")
	}

	// Escape special characters for MarkdownV2
	escapedUName := escapeMarkdown(roomInfo.UName)
	escapedTitle := escapeMarkdown(roomInfo.Title)

	// Build the message
	var message string

	// Main header with emoji and bold anchor name
	message += fmt.Sprintf("🔴 *%s 开始直播啦！*\n\n", escapedUName)

	// Live room title (plain text)
	if roomInfo.Title != "" {
		message += fmt.Sprintf("🎥 直播标题：%s\n\n", escapedTitle)
	} else {
		message += "🎥 直播标题：未设置\n\n"
	}

	// Live start time
	message += fmt.Sprintf("⏰ 开播时间：_%s_\n\n", timeStr)

	// Live room link
	message += fmt.Sprintf("[👉 进入直播间](%s)", liveURL)

	// Determine which image to use (prefer user_cover, fall back to keyframe)
	photoURL := roomInfo.UserCover
	if photoURL == "" && roomInfo.Keyframe != "" {
		photoURL = roomInfo.Keyframe
	}

	return message, photoURL
}

// FormatLiveEndNotification formats a notification for when a live stream ends
func FormatLiveEndNotification(roomInfo models.RoomInfo) string {
	// Use real room ID for links if available
	roomID := roomInfo.RealRoomID
	if roomID == "" {
		roomID = roomInfo.RoomID
	}

	// Escape special characters for MarkdownV2
	escapedUName := escapeMarkdown(roomInfo.UName)

	// Format the end time
	timeStr := ""
	if !roomInfo.EndTime.IsZero() {
		timeStr = roomInfo.EndTime.Format("2006-01-02 15:04:05")
	} else {
		timeStr = time.Now().Format("2006-01-02 15:04:05")
	}

	// Anchor's space URL
	spaceURL := ""
	if roomInfo.UID != "" {
		spaceURL = fmt.Sprintf("https://space.bilibili.com/%s", roomInfo.UID)
	}

	// Build the message
	message := fmt.Sprintf("💤 *%s* 已经下播了\n\n", escapedUName)

	// End time
	message += fmt.Sprintf("⏰ 下播时间：_%s_\n\n", timeStr)

	// Links
	if spaceURL != "" {
		message += fmt.Sprintf("[🏠 主播主页](%s)\n", spaceURL)
	}

	// Add live room link
	message += fmt.Sprintf("[🎬 直播间回放](https://live.bilibili.com/%s)", roomID)

	return message
}

// FormatStatusNotification formats a general status notification
func FormatStatusNotification(status string, details map[string]interface{}) string {
	message := fmt.Sprintf("📊 *状态更新*\n\n%s\n\n", status)

	if len(details) > 0 {
		message += "*详细信息：*\n"
		for key, value := range details {
			message += fmt.Sprintf("• *%s*: `%v`\n", key, value)
		}
	}

	return message
}
