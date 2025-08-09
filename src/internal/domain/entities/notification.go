package entities

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypePlayerChange NotificationType = "player_change"
	NotificationTypeGameEnd      NotificationType = "game_end"
	NotificationTypeTurnStart    NotificationType = "turn_start"
)

// Notification represents a notification to be sent
type Notification struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	GameID    string           `json:"gameId"`
	ChannelID string           `json:"channelId"`
	GuildID   string           `json:"guildId"`
	Message   string           `json:"message"`
	Users     []string         `json:"users"` // Discord user IDs to mention
	Sent      bool             `json:"sent"`
	Attempts  int              `json:"attempts"`
	CreatedAt time.Time        `json:"createdAt"`
	SentAt    *time.Time       `json:"sentAt,omitempty"`
}

// String returns the string representation of the notification type
func (nt NotificationType) String() string {
	return string(nt)
}

// ParseNotificationType parses a string into a NotificationType
func ParseNotificationType(s string) NotificationType {
	return NotificationType(s)
}

// IsSent returns true if the notification has been sent
func (n *Notification) IsSent() bool {
	return n.Sent
}

// GetMessage returns the notification message
func (n *Notification) GetMessage() string {
	return n.Message
}

// GetType returns the notification type
func (n *Notification) GetType() NotificationType {
	return n.Type
}

// GetTimestamp returns the creation timestamp
func (n *Notification) GetTimestamp() time.Time {
	return n.CreatedAt
}

// MarkAsSent marks the notification as sent
func (n *Notification) MarkAsSent() {
	n.Sent = true
	if n.SentAt == nil {
		now := time.Now()
		n.SentAt = &now
	}
}

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	GetNotification(id string) (*Notification, error)
	SaveNotification(notification *Notification) error
	UpdateNotification(notification *Notification) error
	DeleteNotification(id string) error
	GetUnsentNotifications() ([]*Notification, error)
	GetNotificationsByGame(gameID string) ([]*Notification, error)
}
