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
	ready       bool // Track if we've received the READY event
	sequence    *int // Track the sequence number from the last message
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
	Op       int             `json:"op"`
	Data     json.RawMessage `json:"d,omitempty"`
	Type     string          `json:"t,omitempty"`
	Sequence *int            `json:"s,omitempty"`
}

// DiscordIdentifyPayload represents the identify payload for Discord Gateway
type DiscordIdentifyPayload struct {
	Token   string `json:"token"`
	Intents int    `json:"intents"`
}

// NewDiscordAdapter creates a new Discord adapter
func NewDiscordAdapter() *DiscordAdapter {
	return &DiscordAdapter{
		token:       os.Getenv("DISCORD_TOKEN"),
		appID:       os.Getenv("DISCORD_APP_ID"),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		commandChan: make(chan DiscordCommand, 100),
		stopChan:    make(chan struct{}),
		ready:       false,
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

	var gateway struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gateway); err != nil {
		return "", fmt.Errorf("failed to decode gateway response: %w", err)
	}

	return gateway.URL + "?v=10&encoding=json", nil
}

// handleWebSocket handles the WebSocket connection to Discord Gateway
func (a *DiscordAdapter) handleWebSocket() {
	defer a.wsConn.Close()

	// Send identify payload
	if err := a.sendIdentify(); err != nil {
		log.Printf("Failed to identify: %v", err)
		return
	}

	log.Printf("Successfully identified with Discord Gateway")

	for {
		select {
		case <-a.stopChan:
			log.Printf("Stopping WebSocket connection")
			return
		default:
			_, message, err := a.wsConn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)

				// Try to reconnect if it's a connection error
				if strings.Contains(err.Error(), "close") {
					log.Printf("Connection closed, attempting to reconnect...")
					time.Sleep(5 * time.Second)

					// Try to reconnect
					if err := a.reconnect(); err != nil {
						log.Printf("Failed to reconnect: %v", err)
						return
					}
					continue
				}
				return
			}

			if err := a.handleMessage(message); err != nil {
				log.Printf("Failed to handle message: %v", err)
			}
		}
	}
}

// reconnect attempts to reconnect to the Discord Gateway
func (a *DiscordAdapter) reconnect() error {
	log.Printf("Attempting to reconnect to Discord Gateway...")

	// Close existing connection
	if a.wsConn != nil {
		a.wsConn.Close()
	}

	// Reset ready flag
	a.ready = false

	// Get new gateway URL
	gatewayURL, err := a.getGatewayURL()
	if err != nil {
		return fmt.Errorf("failed to get gateway URL: %w", err)
	}

	// Connect to WebSocket
	wsConn, _, err := websocket.DefaultDialer.Dial(gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	a.wsConn = wsConn

	// Send identify payload
	if err := a.sendIdentify(); err != nil {
		return fmt.Errorf("failed to identify after reconnect: %w", err)
	}

	log.Printf("Successfully reconnected to Discord Gateway")
	return nil
}

// sendIdentify sends the identify payload to Discord
func (a *DiscordAdapter) sendIdentify() error {
	identify := DiscordGatewayPayload{
		Op: 2, // Identify
		Data: func() json.RawMessage {
			payload := map[string]interface{}{
				"token":   a.token,
				"intents": 33281, // GUILDS + GUILD_MESSAGES + MESSAGE_CONTENT
				"properties": map[string]string{
					"os":      "linux",
					"browser": "18xxNotifier",
					"device":  "18xxNotifier",
				},
			}

			data, _ := json.Marshal(payload)
			log.Printf("Identify payload JSON: %s", string(data))
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

	// Store sequence number if present
	if payload.Sequence != nil {
		a.sequence = payload.Sequence
		log.Printf("Updated sequence number: %d", *a.sequence)
	}

	switch payload.Op {
	case 10: // Hello - start heartbeat
		go a.startHeartbeat(payload.Data)
	case 11: // Heartbeat acknowledgment
		log.Printf("Received heartbeat acknowledgment")
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

	log.Printf("Starting heartbeat with interval: %dms", hello.HeartbeatInterval)
	ticker := time.NewTicker(time.Duration(hello.HeartbeatInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Sending heartbeat...")

			// Create heartbeat payload with sequence number
			var heartbeatData json.RawMessage
			if a.sequence != nil {
				// Include the sequence number
				data, _ := json.Marshal(*a.sequence)
				heartbeatData = data
			} else {
				// Send null if no sequence number
				heartbeatData = json.RawMessage("null")
			}

			heartbeat := DiscordGatewayPayload{
				Op:   1, // Heartbeat
				Data: heartbeatData,
			}

			if err := a.wsConn.WriteJSON(heartbeat); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
			log.Printf("Heartbeat sent successfully")
		case <-a.stopChan:
			log.Printf("Stopping heartbeat loop")
			return
		}
	}
}

// handleDispatch handles Discord Gateway dispatch events
func (a *DiscordAdapter) handleDispatch(eventType string, data json.RawMessage) error {
	switch eventType {
	case "READY":
		a.ready = true
		log.Printf("Received READY event from Discord Gateway")
	case "MESSAGE_CREATE":
		return a.handleMessageCreate(data)
	}
	return nil
}

// handleMessageCreate processes message creation events
func (a *DiscordAdapter) handleMessageCreate(data json.RawMessage) error {
	// Debug: Log the raw data first
	log.Printf("Raw message data: %s", string(data))

	var message DiscordMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Debug logging
	log.Printf("Received message: '%s' from user %s in channel %s", message.Content, message.Author.ID, message.ChannelID)

	// Ignore messages from bots
	if message.Author.ID == a.appID {
		log.Printf("Ignoring message from self (bot)")
		return nil
	}

	// Only process messages if the bot is ready
	if !a.ready {
		log.Printf("Bot not ready, ignoring message: '%s'", message.Content)
		return nil
	}

	// Check if message content is empty
	if message.Content == "" {
		log.Printf("Message content is empty, trying to extract from raw data...")

		// Try to extract content directly from raw data as a fallback
		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err == nil {
			if content, exists := rawData["content"].(string); exists && content != "" {
				log.Printf("Found content in raw data: '%s'", content)
				message.Content = content
			} else {
				log.Printf("No content found in raw data either")
				return nil
			}
		} else {
			log.Printf("Failed to parse raw data as map: %v", err)
			return nil
		}
	}

	// Check if message is a command (starts with !18xx)
	if strings.HasPrefix(message.Content, "!18xx") {
		log.Printf("Processing command: %s", message.Content)
		parts := strings.Fields(message.Content)
		if len(parts) < 2 {
			log.Printf("Command has insufficient parts: %d", len(parts))
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
	} else {
		log.Printf("Message is not a command (doesn't start with !18xx): '%s'", message.Content)
	}

	return nil
}
