package app

import (
	"fmt"
	"time"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/domain/entities"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	notificationRepo entities.NotificationRepository
	userRepo         entities.UserRepository
	channelRepo      entities.ChannelRepository
	discordAdapter   *adapters.DiscordAdapter
}

// NewNotificationService creates a new NotificationService instance
func NewNotificationService(
	notificationRepo entities.NotificationRepository,
	userRepo entities.UserRepository,
	channelRepo entities.ChannelRepository,
	discordAdapter *adapters.DiscordAdapter,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		channelRepo:      channelRepo,
		discordAdapter:   discordAdapter,
	}
}

// HandleGameChange processes a game change and creates notifications
func (s *NotificationService) HandleGameChange(change *entities.GameChange) error {
	// Get users subscribed to this game
	users, err := s.userRepo.GetUsersByGame(change.GameID)
	if err != nil {
		return fmt.Errorf("failed to get users for game: %w", err)
	}

	// Get channels tracking this game
	channels, err := s.channelRepo.GetChannelsByGame(change.GameID)
	if err != nil {
		return fmt.Errorf("failed to get channels for game: %w", err)
	}

	// Create notifications for each channel
	for _, channel := range channels {
		// Find users to mention (users with accounts in this game)
		var mentions []string
		for _, user := range users {
			// Check if user has an account that might be involved in this change
			if change.ChangeType == "player_change" {
				if user.HasAccount(change.OldValue) || user.HasAccount(change.NewValue) {
					mentions = append(mentions, user.DiscordID)
				}
			} else {
				// For other changes, mention all subscribed users
				mentions = append(mentions, user.DiscordID)
			}
		}

		// Create notification message
		message := s.createNotificationMessage(change)

		// Create notification
		notification := &entities.Notification{
			ID:        fmt.Sprintf("notif_%s_%s_%d", change.GameID, channel.ID, time.Now().Unix()),
			Type:      s.mapChangeTypeToNotificationType(change.ChangeType),
			Message:   message,
			GameID:    change.GameID,
			ChannelID: channel.ID,
			GuildID:   channel.GuildID,
			Users:     mentions,
			Sent:      false,
			CreatedAt: time.Now(),
		}

		// Save notification
		err = s.notificationRepo.SaveNotification(notification)
		if err != nil {
			return fmt.Errorf("failed to save notification: %w", err)
		}
	}

	return nil
}

// SendPendingNotifications sends all pending notifications
func (s *NotificationService) SendPendingNotifications() error {
	// Get pending notifications
	notifications, err := s.notificationRepo.GetUnsentNotifications()
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	// Send each notification
	for _, notification := range notifications {
		// Convert user IDs to mentions
		var mentions []string
		for _, userID := range notification.Users {
			mentions = append(mentions, fmt.Sprintf("<@%s>", userID))
		}

		// Send via Discord
		err = s.discordAdapter.SendNotification(notification.ChannelID, notification.Message, mentions)
		if err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		// Mark as sent
		notification.MarkAsSent()
		err = s.notificationRepo.UpdateNotification(notification)
		if err != nil {
			return fmt.Errorf("failed to mark notification as sent: %w", err)
		}
	}

	return nil
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(gameID, channelID, message string, notificationType entities.NotificationType) error {
	notification := &entities.Notification{
		ID:        fmt.Sprintf("notif_%s_%s_%d", gameID, channelID, time.Now().Unix()),
		Type:      notificationType,
		Message:   message,
		GameID:    gameID,
		ChannelID: channelID,
		Users:     []string{},
		Sent:      false,
		CreatedAt: time.Now(),
	}

	return s.notificationRepo.SaveNotification(notification)
}

// GetPendingNotifications returns all pending notifications
func (s *NotificationService) GetPendingNotifications() ([]*entities.Notification, error) {
	return s.notificationRepo.GetUnsentNotifications()
}

// MarkNotificationSent marks a notification as sent
func (s *NotificationService) MarkNotificationSent(notificationID string) error {
	notification, err := s.notificationRepo.GetNotification(notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification == nil {
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	notification.MarkAsSent()
	return s.notificationRepo.UpdateNotification(notification)
}

// createNotificationMessage creates a notification message based on the game change
func (s *NotificationService) createNotificationMessage(change *entities.GameChange) string {
	switch change.ChangeType {
	case "player_change":
		return fmt.Sprintf("🎮 **Game Update**: It's now %s's turn in game %s!", change.NewValue, change.GameID)
	case "game_end":
		return fmt.Sprintf("🏁 **Game Over**: Game %s has ended!", change.GameID)
	case "turn_start":
		return fmt.Sprintf("🚀 **Game Started**: Game %s is now active!", change.GameID)
	default:
		return fmt.Sprintf("📢 **Game Update**: %s in game %s", change.ChangeType, change.GameID)
	}
}

// mapChangeTypeToNotificationType maps game change types to notification types
func (s *NotificationService) mapChangeTypeToNotificationType(changeType string) entities.NotificationType {
	switch changeType {
	case "player_change":
		return entities.NotificationTypePlayerChange
	case "game_end":
		return entities.NotificationTypeGameEnd
	case "turn_start":
		return entities.NotificationTypeTurnStart
	default:
		return entities.NotificationTypePlayerChange
	}
}
