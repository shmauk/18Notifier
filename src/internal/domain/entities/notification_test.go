package entities

import (
	"testing"
	"time"
)

func TestNotificationType_String(t *testing.T) {
	tests := []struct {
		name      string
		notifType NotificationType
		expected  string
	}{
		{
			name:      "player change",
			notifType: NotificationTypePlayerChange,
			expected:  "player_change",
		},
		{
			name:      "game end",
			notifType: NotificationTypeGameEnd,
			expected:  "game_end",
		},
		{
			name:      "turn start",
			notifType: NotificationTypeTurnStart,
			expected:  "turn_start",
		},
		{
			name:      "unknown type",
			notifType: NotificationType("unknown"),
			expected:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.notifType.String()
			if result != tt.expected {
				t.Errorf("NotificationType.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestParseNotificationType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NotificationType
	}{
		{
			name:     "player change",
			input:    "player_change",
			expected: NotificationTypePlayerChange,
		},
		{
			name:     "game end",
			input:    "game_end",
			expected: NotificationTypeGameEnd,
		},
		{
			name:     "turn start",
			input:    "turn_start",
			expected: NotificationTypeTurnStart,
		},
		{
			name:     "unknown type",
			input:    "unknown",
			expected: NotificationType("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNotificationType(tt.input)
			if result != tt.expected {
				t.Errorf("ParseNotificationType(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNotification_IsSent(t *testing.T) {
	tests := []struct {
		name     string
		notif    *Notification
		expected bool
	}{
		{
			name: "sent notification",
			notif: &Notification{
				ID:        "notif1",
				Type:      NotificationTypePlayerChange,
				Message:   "Player changed",
				GameID:    "game1",
				ChannelID: "channel1",
				GuildID:   "guild1",
				Users:     []string{"user1"},
				Sent:      true,
				CreatedAt: time.Now(),
			},
			expected: true,
		},
		{
			name: "unsent notification",
			notif: &Notification{
				ID:        "notif2",
				Type:      NotificationTypeGameEnd,
				Message:   "Game ended",
				GameID:    "game1",
				ChannelID: "channel1",
				GuildID:   "guild1",
				Users:     []string{"user1"},
				Sent:      false,
				CreatedAt: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.notif.IsSent()
			if result != tt.expected {
				t.Errorf("Notification.IsSent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNotification_GetMessage(t *testing.T) {
	notif := &Notification{
		ID:        "notif1",
		Type:      NotificationTypePlayerChange,
		Message:   "Player changed from player1 to player2",
		GameID:    "game1",
		ChannelID: "channel1",
		GuildID:   "guild1",
		Users:     []string{"user1"},
		CreatedAt: time.Now(),
		Sent:      false,
	}

	result := notif.GetMessage()
	expected := "Player changed from player1 to player2"

	if result != expected {
		t.Errorf("Notification.GetMessage() = %s, want %s", result, expected)
	}
}

func TestNotification_GetType(t *testing.T) {
	notif := &Notification{
		ID:        "notif1",
		Type:      NotificationTypePlayerChange,
		Message:   "Player changed",
		GameID:    "game1",
		ChannelID: "channel1",
		GuildID:   "guild1",
		Users:     []string{"user1"},
		CreatedAt: time.Now(),
		Sent:      false,
	}

	result := notif.GetType()
	expected := NotificationTypePlayerChange

	if result != expected {
		t.Errorf("Notification.GetType() = %v, want %v", result, expected)
	}
}

func TestNotification_GetTimestamp(t *testing.T) {
	now := time.Now()
	notif := &Notification{
		ID:        "notif1",
		Type:      NotificationTypePlayerChange,
		Message:   "Player changed",
		GameID:    "game1",
		ChannelID: "channel1",
		GuildID:   "guild1",
		Users:     []string{"user1"},
		CreatedAt: now,
		Sent:      false,
	}

	result := notif.GetTimestamp()
	if !result.Equal(now) {
		t.Errorf("Notification.GetTimestamp() = %v, want %v", result, now)
	}
}

func TestNotification_MarkAsSent(t *testing.T) {
	notif := &Notification{
		ID:        "notif1",
		Type:      NotificationTypePlayerChange,
		Message:   "Player changed",
		GameID:    "game1",
		ChannelID: "channel1",
		GuildID:   "guild1",
		Users:     []string{"user1"},
		CreatedAt: time.Now(),
		Sent:      false,
	}

	notif.MarkAsSent()

	if !notif.Sent {
		t.Errorf("Notification.MarkAsSent() did not mark notification as sent")
	}
}
