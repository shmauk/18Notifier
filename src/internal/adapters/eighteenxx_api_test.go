package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEighteenxxAPIAdapter_FetchGameData(t *testing.T) {
	// Create a test server to mock the 18xx.games API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Extract game ID from URL path (e.g., /api/game/test-game-123)
		pathParts := strings.Split(r.URL.Path, "/")
		var gameID string
		if len(pathParts) >= 3 {
			gameID = pathParts[len(pathParts)-1]
		}

		var response map[string]interface{}

		switch gameID {
		case "test-game-123":
			response = map[string]interface{}{
				"id":           "test-game-123",
				"players":      []string{"player1", "player2", "player3"},
				"activePlayer": "player1",
				"finished":     false,
				"lastUpdated":  time.Now().Format(time.RFC3339),
			}
		case "finished-game":
			response = map[string]interface{}{
				"id":           "finished-game",
				"players":      []string{"player1", "player2"},
				"activePlayer": "",
				"finished":     true,
				"lastUpdated":  time.Now().Format(time.RFC3339),
			}
		case "not-found":
			w.WriteHeader(http.StatusNotFound)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create adapter with test server URL
	adapter := &EighteenxxAPIAdapter{
		baseURL:    server.URL + "/api/game",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	tests := []struct {
		name    string
		gameID  string
		wantErr bool
		wantNil bool
	}{
		{
			name:    "successful game fetch",
			gameID:  "test-game-123",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "finished game fetch",
			gameID:  "finished-game",
			wantErr: false,
			wantNil: false,
		},
		{
			name:    "game not found",
			gameID:  "not-found",
			wantErr: true,
			wantNil: true,
		},
		{
			name:    "server error",
			gameID:  "server-error",
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game, err := adapter.FetchGameData(tt.gameID)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if tt.wantNil && game != nil {
				t.Errorf("Expected nil game but got: %+v", game)
			}
			if !tt.wantNil && game == nil {
				t.Errorf("Expected game but got nil")
			}

			// Verify game data for successful cases
			if game != nil {
				if game.ID != tt.gameID {
					t.Errorf("Expected game ID %s, got %s", tt.gameID, game.ID)
				}
				if tt.gameID == "finished-game" && !game.Finished {
					t.Errorf("Expected finished game to be marked as finished")
				}
				if tt.gameID == "test-game-123" && game.Finished {
					t.Errorf("Expected active game to not be marked as finished")
				}
			}
		})
	}
}

func TestEighteenxxAPIAdapter_NewEighteenxxAPIAdapter(t *testing.T) {
	adapter := NewEighteenxxAPIAdapter()

	if adapter == nil {
		t.Error("Expected adapter to be created, got nil")
	}

	if adapter.baseURL != "https://18xx.games/api/game" {
		t.Errorf("Expected base URL 'https://18xx.games/api/game', got '%s'", adapter.baseURL)
	}

	if adapter.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}
