package app

import (
	"log"
	"time"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/domain/entities"
)

// NotificationHandler handles notification events and processing
type NotificationHandler struct {
	notificationService *NotificationService
	discordAdapter      *adapters.DiscordAdapter
	eventChan           chan *entities.GameChange
	stopChan            chan struct{}
}

// NewNotificationHandler creates a new NotificationHandler instance
func NewNotificationHandler(notificationService *NotificationService, discordAdapter *adapters.DiscordAdapter) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		discordAdapter:      discordAdapter,
		eventChan:           make(chan *entities.GameChange, 100),
		stopChan:            make(chan struct{}),
	}
}

// Start starts the notification handler
func (h *NotificationHandler) Start() {
	go h.processEvents()
	log.Println("NotificationHandler started")
}

// Stop stops the notification handler
func (h *NotificationHandler) Stop() {
	close(h.stopChan)
	log.Println("NotificationHandler stopped")
}

// HandleGameChange handles a game change event
func (h *NotificationHandler) HandleGameChange(change *entities.GameChange) {
	select {
	case h.eventChan <- change:
		// Event queued successfully
	default:
		log.Printf("Warning: Event channel full, dropping game change for game %s", change.GameID)
	}
}

// SendTestNotification sends a test notification to a channel
func (h *NotificationHandler) SendTestNotification(channelID, message string) error {
	return h.discordAdapter.SendNotification(channelID, message, []string{})
}

// processEvents processes game change events
func (h *NotificationHandler) processEvents() {
	ticker := time.NewTicker(30 * time.Second) // Send notifications every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case change := <-h.eventChan:
			h.processGameChange(change)
		case <-ticker.C:
			// Send pending notifications periodically
			err := h.notificationService.SendPendingNotifications()
			if err != nil {
				log.Printf("Error sending pending notifications: %v", err)
			}
		case <-h.stopChan:
			return
		}
	}
}

// processGameChange processes a single game change
func (h *NotificationHandler) processGameChange(change *entities.GameChange) {
	// Handle the game change
	err := h.notificationService.HandleGameChange(change)
	if err != nil {
		log.Printf("Error handling game change: %v", err)
		return
	}

	// Send pending notifications immediately for important changes
	if change.ChangeType == "player_change" || change.ChangeType == "game_end" {
		err = h.notificationService.SendPendingNotifications()
		if err != nil {
			log.Printf("Error sending immediate notifications: %v", err)
		}
	}
}
