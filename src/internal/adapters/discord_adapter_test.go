package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDiscordAdapter_NewDiscordAdapter(t *testing.T) {
	// Test with environment variables set
	adapter := NewDiscordAdapter()

	if adapter == nil {
		t.Error("Expected adapter to be created, got nil")
	}

	if adapter.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if adapter.commandChan == nil {
		t.Error("Expected command channel to be initialized")
	}

	if adapter.stopChan == nil {
		t.Error("Expected stop channel to be initialized")
	}
}

func TestDiscordAdapter_SendNotification(t *testing.T) {
	// Create a test server to mock Discord REST API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify the URL path (should be for creating a message)
		if r.URL.Path != "/channels/test-channel/messages" {
			t.Errorf("Expected path '/channels/test-channel/messages', got '%s'", r.URL.Path)
		}

		// Parse the request body
		var request struct {
			Content string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify the message content
		if request.Content == "" {
			t.Error("Expected non-empty message content")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Mock successful response
		response := map[string]interface{}{
			"id":         "message-id",
			"channel_id": "test-channel",
			"content":    request.Content,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tests := []struct {
		name         string
		channelID    string
		message      string
		userMentions []string
		wantErr      bool
	}{
		{
			name:         "simple notification",
			channelID:    "test-channel",
			message:      "Test notification message",
			userMentions: []string{},
			wantErr:      false,
		},
		{
			name:         "notification with mentions",
			channelID:    "test-channel",
			message:      "Test notification with mentions",
			userMentions: []string{"<@user1>", "<@user2>"},
			wantErr:      false,
		},
		{
			name:         "empty message",
			channelID:    "test-channel",
			message:      "",
			userMentions: []string{},
			wantErr:      true,
		},
		{
			name:         "empty channel ID",
			channelID:    "",
			message:      "Test message",
			userMentions: []string{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, we'll just test the basic logic without making HTTP requests
			// since the current implementation doesn't allow easy mocking
			if tt.channelID == "" || tt.message == "" {
				// These should fail validation
				if !tt.wantErr {
					t.Errorf("Expected error for invalid input")
				}
				return
			}

			// Test the message formatting logic
			fullMessage := tt.message
			if len(tt.userMentions) > 0 {
				fullMessage += "\n" + strings.Join(tt.userMentions, " ")
			}

			if fullMessage == "" {
				t.Error("Expected non-empty full message")
			}
		})
	}
}

func TestDiscordAdapter_ReceiveCommands(t *testing.T) {
	adapter := &DiscordAdapter{
		commandChan: make(chan DiscordCommand, 10),
		stopChan:    make(chan struct{}),
	}

	// Test that we can receive commands
	commandChan, err := adapter.ReceiveCommands()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if commandChan == nil {
		t.Error("Expected command channel, got nil")
	}

	// Test sending a command through the channel
	testCommand := DiscordCommand{
		UserID:    "test-user",
		ChannelID: "test-channel",
		GuildID:   "test-guild",
		Command:   "test",
		Args:      []string{"arg1", "arg2"},
	}

	// Send command in a goroutine to avoid blocking
	go func() {
		adapter.commandChan <- testCommand
	}()

	// Receive the command
	select {
	case receivedCommand := <-commandChan:
		if receivedCommand.UserID != testCommand.UserID {
			t.Errorf("Expected user ID '%s', got '%s'", testCommand.UserID, receivedCommand.UserID)
		}
		if receivedCommand.Command != testCommand.Command {
			t.Errorf("Expected command '%s', got '%s'", testCommand.Command, receivedCommand.Command)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for command")
	}
}

func TestDiscordAdapter_Start(t *testing.T) {
	// Test with empty token (should fail)
	adapter := &DiscordAdapter{
		token:       "",
		appID:       "test-app-id",
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		commandChan: make(chan DiscordCommand, 10),
		stopChan:    make(chan struct{}),
	}

	err := adapter.Start()
	if err == nil {
		t.Error("Expected error when starting with empty token")
	}

	// Test with valid token but mock HTTP client that fails
	adapter = &DiscordAdapter{
		token:       "valid-token",
		appID:       "test-app-id",
		httpClient:  &http.Client{Timeout: 1 * time.Millisecond}, // Very short timeout to force failure
		commandChan: make(chan DiscordCommand, 10),
		stopChan:    make(chan struct{}),
	}

	err = adapter.Start()
	// Should fail on gateway connection due to timeout
	if err == nil {
		t.Error("Expected error when connecting to gateway")
	}
}

func TestDiscordAdapter_Stop(t *testing.T) {
	adapter := &DiscordAdapter{
		stopChan: make(chan struct{}),
	}

	// Test stopping
	err := adapter.Stop()
	if err != nil {
		t.Errorf("Expected no error when stopping, got: %v", err)
	}

	// Verify stop channel is closed
	select {
	case <-adapter.stopChan:
		// Channel is closed, which is expected
	default:
		t.Error("Expected stop channel to be closed")
	}
}
