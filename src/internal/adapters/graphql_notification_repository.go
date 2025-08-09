package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/18xxnotifier/internal/domain/entities"
)

// GraphQLNotificationRepository implements entities.NotificationRepository using GraphQL
type GraphQLNotificationRepository struct {
	graphqlAdapter GraphQLAdapter
}

// NewGraphQLNotificationRepository creates a new GraphQL notification repository
func NewGraphQLNotificationRepository(graphqlAdapter GraphQLAdapter) *GraphQLNotificationRepository {
	return &GraphQLNotificationRepository{
		graphqlAdapter: graphqlAdapter,
	}
}

// SaveNotification saves a new notification
func (r *GraphQLNotificationRepository) SaveNotification(notification *entities.Notification) error {
	mutation := `
		mutation AddNotification($notification: AddNotificationInput!) {
			addNotification(input: [$notification]) {
				notification {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"type":      notification.Type.String(),
		"message":   notification.Message,
		"gameId":    notification.GameID,
		"channelId": notification.ChannelID,
		"guildId":   notification.GuildID,
		"users":     notification.Users,
		"sent":      notification.Sent,
		"attempts":  notification.Attempts,
		"createdAt": notification.CreatedAt,
	}

	variables := map[string]interface{}{
		"notification": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	return nil
}

// GetNotification retrieves a notification by ID
func (r *GraphQLNotificationRepository) GetNotification(id string) (*entities.Notification, error) {
	query := `
		query GetNotification($id: String!) {
			queryNotification(filter: { id: { eq: $id } }) {
				id
				type
				message
				gameId
				channelId
				guildId
				users
				sent
				attempts
				createdAt
			}
		}
	`

	variables := map[string]interface{}{
		"id": id,
	}

	result, err := r.graphqlAdapter.Query(query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification: %w", err)
	}

	var response struct {
		Data struct {
			QueryNotification []struct {
				ID        string   `json:"id"`
				Type      string   `json:"type"`
				Message   string   `json:"message"`
				GameID    string   `json:"gameId"`
				ChannelID string   `json:"channelId"`
				GuildID   string   `json:"guildId"`
				Users     []string `json:"users"`
				Sent      bool     `json:"sent"`
				Attempts  int      `json:"attempts"`
				CreatedAt string   `json:"createdAt"`
			} `json:"queryNotification"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Data.QueryNotification) == 0 {
		return nil, fmt.Errorf("notification not found: %s", id)
	}

	notifData := response.Data.QueryNotification[0]
	notificationType := entities.ParseNotificationType(notifData.Type)
	notification := &entities.Notification{
		ID:        notifData.ID,
		Type:      notificationType,
		Message:   notifData.Message,
		GameID:    notifData.GameID,
		ChannelID: notifData.ChannelID,
		GuildID:   notifData.GuildID,
		Users:     notifData.Users,
		Sent:      notifData.Sent,
		Attempts:  notifData.Attempts,
	}

	return notification, nil
}

// UpdateNotification updates an existing notification
func (r *GraphQLNotificationRepository) UpdateNotification(notification *entities.Notification) error {
	mutation := `
		mutation UpdateNotification($notification: UpdateNotificationInput!) {
			updateNotification(input: $notification) {
				notification {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"id": map[string]interface{}{
				"eq": notification.ID,
			},
		},
		"set": map[string]interface{}{
			"type":      notification.Type.String(),
			"message":   notification.Message,
			"gameId":    notification.GameID,
			"channelId": notification.ChannelID,
			"guildId":   notification.GuildID,
			"users":     notification.Users,
			"sent":      notification.Sent,
			"attempts":  notification.Attempts,
		},
	}

	variables := map[string]interface{}{
		"notification": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

// GetPendingNotifications gets all unsent notifications
func (r *GraphQLNotificationRepository) GetPendingNotifications() ([]*entities.Notification, error) {
	query := `
		query GetPendingNotifications {
			queryNotification {
				id
				type
				message
				gameId
				channelId
				guildId
				users
				sent
				attempts
				createdAt
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending notifications: %w", err)
	}

	var response struct {
		Data struct {
			QueryNotification []struct {
				ID        string   `json:"id"`
				Type      string   `json:"type"`
				Message   string   `json:"message"`
				GameID    string   `json:"gameId"`
				ChannelID string   `json:"channelId"`
				GuildID   string   `json:"guildId"`
				Users     []string `json:"users"`
				Sent      bool     `json:"sent"`
				Attempts  int      `json:"attempts"`
				CreatedAt string   `json:"createdAt"`
			} `json:"queryNotification"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var notifications []*entities.Notification
	for _, notifData := range response.Data.QueryNotification {
		// Filter for unsent notifications in application code
		if notifData.Sent {
			continue
		}

		notificationType := entities.ParseNotificationType(notifData.Type)
		notification := &entities.Notification{
			ID:        notifData.ID,
			Type:      notificationType,
			Message:   notifData.Message,
			GameID:    notifData.GameID,
			ChannelID: notifData.ChannelID,
			GuildID:   notifData.GuildID,
			Users:     notifData.Users,
			Sent:      notifData.Sent,
			Attempts:  notifData.Attempts,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnsentNotifications gets all unsent notifications (alias for GetPendingNotifications)
func (r *GraphQLNotificationRepository) GetUnsentNotifications() ([]*entities.Notification, error) {
	return r.GetPendingNotifications()
}

// MarkNotificationSent marks a notification as sent
func (r *GraphQLNotificationRepository) MarkNotificationSent(notificationID string) error {
	mutation := `
		mutation UpdateNotification($notification: UpdateNotificationInput!) {
			updateNotification(input: $notification) {
				notification {
					id
				}
			}
		}
	`

	input := map[string]interface{}{
		"filter": map[string]interface{}{
			"id": map[string]interface{}{
				"eq": notificationID,
			},
		},
		"set": map[string]interface{}{
			"sent": true,
		},
	}

	variables := map[string]interface{}{
		"notification": input,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return nil
}

// GetNotificationsByGame gets all notifications for a specific game
func (r *GraphQLNotificationRepository) GetNotificationsByGame(gameID string) ([]*entities.Notification, error) {
	query := `
		query GetNotificationsByGame {
			queryNotification {
				id
				type
				message
				gameId
				channelId
				guildId
				users
				sent
				attempts
				createdAt
			}
		}
	`

	result, err := r.graphqlAdapter.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications by game: %w", err)
	}

	var response struct {
		Data struct {
			QueryNotification []struct {
				ID        string   `json:"id"`
				Type      string   `json:"type"`
				Message   string   `json:"message"`
				GameID    string   `json:"gameId"`
				ChannelID string   `json:"channelId"`
				GuildID   string   `json:"guildId"`
				Users     []string `json:"users"`
				Sent      bool     `json:"sent"`
				Attempts  int      `json:"attempts"`
				CreatedAt string   `json:"createdAt"`
			} `json:"queryNotification"`
		} `json:"data"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var notifications []*entities.Notification
	for _, notifData := range response.Data.QueryNotification {
		// Filter for notifications by game ID in application code
		if notifData.GameID != gameID {
			continue
		}

		notificationType := entities.ParseNotificationType(notifData.Type)
		notification := &entities.Notification{
			ID:        notifData.ID,
			Type:      notificationType,
			Message:   notifData.Message,
			GameID:    notifData.GameID,
			ChannelID: notifData.ChannelID,
			GuildID:   notifData.GuildID,
			Users:     notifData.Users,
			Sent:      notifData.Sent,
			Attempts:  notifData.Attempts,
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// DeleteNotification deletes a notification by ID
func (r *GraphQLNotificationRepository) DeleteNotification(notificationID string) error {
	mutation := `
		mutation DeleteNotification($filter: NotificationFilter!) {
			deleteNotification(filter: $filter) {
				msg
			}
		}
	`

	filter := map[string]interface{}{
		"id": map[string]interface{}{
			"eq": notificationID,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	_, err := r.graphqlAdapter.Mutate(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}
