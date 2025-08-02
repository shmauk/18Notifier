package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// DiscordAdapter implements DiscordAdapter interface for Discord bot functionality
type DiscordAdapter struct {
	token       string
	appID       string
	wsConn      *websocket.Conn
	httpClient  *http.Client
	commandChan chan DiscordCommand
	stopChan    chan struct{}
}

// DiscordMessage represents a Discord message structure
type DiscordMessage struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id"`
	Author    struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"author"`
	Content string `json:"content"`
}

// DiscordGatewayPayload represents Discord Gateway payload
type DiscordGatewayPayload struct {
	Op   int             `json:"op"`
	Data json.RawMessage `json:"d,omitempty"`
	Type string          `json:"t,omitempty"`
}

// DiscordIdentifyPayload represents the identify payload for Discord Gateway
type DiscordIdentifyPayload struct {
	Token      string `json:"token"`
	Properties struct {
		OS      string `json:"os"`
		Browser string `json:"browser"`
		Device  string `json:"device"`
	} `json:"properties"`
	Presence struct {
		Status string `json:"status"`
		Since  int    `json:"since"`
		Game   struct {
			Name string `json:"name"`
			Type int    `json:"type"`
		} `json:"game"`
	} `json:"presence"`
}

// NewDiscordAdapter creates a new Discord adapter
func NewDiscordAdapter() *DiscordAdapter {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Println("Warning: DISCORD_TOKEN not set")
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		log.Println("Warning: DISCORD_APP_ID not set")
	}

	return &DiscordAdapter{
		token:       token,
		appID:       appID,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		commandChan: make(chan DiscordCommand, 100),
		stopChan:    make(chan struct{}),
	}
}

// Start connects to Discord Gateway and begins listening for events
func (a *DiscordAdapter) Start() error {
	if a.token == "" {
		return fmt.Errorf("Discord token not configured")
	}

	// Get Gateway URL
	gatewayURL, err := a.getGatewayURL()
	if err != nil {
		return fmt.Errorf("failed to get gateway URL: %w", err)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Discord Gateway: %w", err)
	}
	a.wsConn = conn

	// Start heartbeat and event handling
	go a.handleWebSocket()

	log.Println("Discord adapter started successfully")
	return nil
}

// Stop disconnects from Discord Gateway
func (a *DiscordAdapter) Stop() error {
	// Always close the stop channel to signal shutdown
	close(a.stopChan)

	if a.wsConn != nil {
		return a.wsConn.Close()
	}
	return nil
}

// SendNotification sends a notification to a Discord channel
func (a *DiscordAdapter) SendNotification(channelID string, message string, userMentions []string) error {
	// Build message with mentions
	var content strings.Builder
	for _, userID := range userMentions {
		content.WriteString(fmt.Sprintf("<@%s> ", userID))
	}
	content.WriteString(message)

	payload := map[string]interface{}{
		"content": content.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+a.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status: %d", resp.StatusCode)
	}

	log.Printf("Sent notification to channel %s: %s", channelID, message)
	return nil
}

// ReceiveCommands returns a channel for receiving Discord commands
func (a *DiscordAdapter) ReceiveCommands() (<-chan DiscordCommand, error) {
	return a.commandChan, nil
}

// getGatewayURL retrieves the Discord Gateway URL
func (a *DiscordAdapter) getGatewayURL() (string, error) {
	resp, err := a.httpClient.Get("https://discord.com/api/v10/gateway")
	if err != nil {
		return "", fmt.Errorf("failed to get gateway URL: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode gateway response: %w", err)
	}

	return result.URL + "?v=10&encoding=json", nil
}

// handleWebSocket handles the WebSocket connection to Discord Gateway
func (a *DiscordAdapter) handleWebSocket() {
	defer a.wsConn.Close()

	// Send identify payload
	if err := a.sendIdentify(); err != nil {
		log.Printf("Failed to identify: %v", err)
		return
	}

	for {
		select {
		case <-a.stopChan:
			return
		default:
			_, message, err := a.wsConn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			if err := a.handleMessage(message); err != nil {
				log.Printf("Failed to handle message: %v", err)
			}
		}
	}
}

// sendIdentify sends the identify payload to Discord
func (a *DiscordAdapter) sendIdentify() error {
	identify := DiscordGatewayPayload{
		Op: 2, // Identify
		Data: func() json.RawMessage {
			payload := DiscordIdentifyPayload{
				Token: a.token,
			}
			payload.Properties.OS = "linux"
			payload.Properties.Browser = "18xxNotifier"
			payload.Properties.Device = "18xxNotifier"
			payload.Presence.Status = "online"
			payload.Presence.Game.Name = "18xx games"
			payload.Presence.Game.Type = 0

			data, _ := json.Marshal(payload)
			return data
		}(),
	}

	return a.wsConn.WriteJSON(identify)
}

// handleMessage processes incoming Discord Gateway messages
func (a *DiscordAdapter) handleMessage(message []byte) error {
	var payload DiscordGatewayPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	switch payload.Op {
	case 10: // Hello - start heartbeat
		go a.startHeartbeat(payload.Data)
	case 0: // Dispatch - handle events
		return a.handleDispatch(payload.Type, payload.Data)
	}

	return nil
}

// startHeartbeat starts the heartbeat loop
func (a *DiscordAdapter) startHeartbeat(data json.RawMessage) {
	var hello struct {
		HeartbeatInterval int `json:"heartbeat_interval"`
	}
	if err := json.Unmarshal(data, &hello); err != nil {
		log.Printf("Failed to parse hello payload: %v", err)
		return
	}

	ticker := time.NewTicker(time.Duration(hello.HeartbeatInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			heartbeat := DiscordGatewayPayload{Op: 1} // Heartbeat
			if err := a.wsConn.WriteJSON(heartbeat); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
		case <-a.stopChan:
			return
		}
	}
}

// handleDispatch handles Discord Gateway dispatch events
func (a *DiscordAdapter) handleDispatch(eventType string, data json.RawMessage) error {
	switch eventType {
	case "MESSAGE_CREATE":
		return a.handleMessageCreate(data)
	}
	return nil
}

// handleMessageCreate processes message creation events
func (a *DiscordAdapter) handleMessageCreate(data json.RawMessage) error {
	var message DiscordMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Ignore messages from bots
	if message.Author.ID == a.appID {
		return nil
	}

	// Check if message is a command
	if strings.HasPrefix(message.Content, "!18xx") {
		parts := strings.Fields(message.Content)
		if len(parts) < 2 {
			return nil
		}

		command := DiscordCommand{
			UserID:    message.Author.ID,
			ChannelID: message.ChannelID,
			GuildID:   message.GuildID,
			Command:   parts[1],
			Args:      parts[2:],
		}

		// Send command to channel
		select {
		case a.commandChan <- command:
			log.Printf("Received command: %s from user %s in channel %s",
				command.Command, command.UserID, command.ChannelID)
		default:
			log.Printf("Command channel full, dropping command: %s", command.Command)
		}
	}

	return nil
}
