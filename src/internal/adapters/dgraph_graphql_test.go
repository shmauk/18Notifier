package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDGraphGraphQLAdapter_Query(t *testing.T) {
	// Create a test server to mock DGraph
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse the request body
		var request struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Mock responses based on the query
		var response interface{}
		switch request.Query {
		case "query HealthCheck { __schema { types { name } } }":
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []map[string]interface{}{
							{"name": "Game"},
							{"name": "User"},
							{"name": "Notification"},
						},
					},
				},
			}
		case "query GetGames { queryGame { id activePlayer finished } }":
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"queryGame": []map[string]interface{}{
						{
							"id":           "game1",
							"activePlayer": "player1",
							"finished":     false,
						},
						{
							"id":           "game2",
							"activePlayer": "",
							"finished":     true,
						},
					},
				},
			}
		default:
			response = map[string]interface{}{
				"data": map[string]interface{}{},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create adapter with test server URL
	adapter := NewDGraphGraphQLAdapter(server.URL)

	tests := []struct {
		name      string
		query     string
		variables map[string]interface{}
		wantErr   bool
	}{
		{
			name:      "health check query",
			query:     "query HealthCheck { __schema { types { name } } }",
			variables: nil,
			wantErr:   false,
		},
		{
			name:      "get games query",
			query:     "query GetGames { queryGame { id activePlayer finished } }",
			variables: nil,
			wantErr:   false,
		},
		{
			name:      "query with variables",
			query:     "query GetGame($id: String!) { getGame(id: $id) { id activePlayer } }",
			variables: map[string]interface{}{"id": "game1"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.Query(tt.query, tt.variables)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Expected result but got nil")
			}

			// Verify the response structure
			if result != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(result, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				if _, exists := response["data"]; !exists {
					t.Error("Expected 'data' field in response")
				}
			}
		})
	}
}

func TestDGraphGraphQLAdapter_Mutate(t *testing.T) {
	// Create a test server to mock DGraph
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse the request body
		var request struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Mock responses based on the mutation
		var response interface{}
		switch request.Query {
		case "mutation AddGame($game: AddGameInput!) { addGame(input: $game) { game { id } } }":
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"addGame": map[string]interface{}{
						"game": map[string]interface{}{
							"id": "new-game-id",
						},
					},
				},
			}
		case "mutation UpdateGame($game: UpdateGameInput!) { updateGame(input: $game) { game { id } } }":
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"updateGame": map[string]interface{}{
						"game": map[string]interface{}{
							"id": "updated-game-id",
						},
					},
				},
			}
		default:
			response = map[string]interface{}{
				"data": map[string]interface{}{},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create adapter with test server URL
	adapter := NewDGraphGraphQLAdapter(server.URL)

	tests := []struct {
		name      string
		mutation  string
		variables map[string]interface{}
		wantErr   bool
	}{
		{
			name:     "add game mutation",
			mutation: "mutation AddGame($game: AddGameInput!) { addGame(input: $game) { game { id } } }",
			variables: map[string]interface{}{
				"game": map[string]interface{}{
					"id":           "new-game",
					"activePlayer": "player1",
					"finished":     false,
				},
			},
			wantErr: false,
		},
		{
			name:     "update game mutation",
			mutation: "mutation UpdateGame($game: UpdateGameInput!) { updateGame(input: $game) { game { id } } }",
			variables: map[string]interface{}{
				"game": map[string]interface{}{
					"filter": map[string]interface{}{
						"id": map[string]interface{}{
							"eq": "game1",
						},
					},
					"set": map[string]interface{}{
						"activePlayer": "player2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.Mutate(tt.mutation, tt.variables)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.wantErr && result == nil {
				t.Errorf("Expected result but got nil")
			}

			// Verify the response structure
			if result != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(result, &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				if _, exists := response["data"]; !exists {
					t.Error("Expected 'data' field in response")
				}
			}
		})
	}
}

func TestDGraphGraphQLAdapter_HealthCheck(t *testing.T) {
	// Create a test server that returns different responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to check if it's a health check
		var request struct {
			Query string `json:"query"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check if it's a health check query (more flexible matching)
		if strings.Contains(request.Query, "HealthCheck") && strings.Contains(request.Query, "__schema") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []map[string]interface{}{
							{"name": "Game"},
							{"name": "User"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	adapter := NewDGraphGraphQLAdapter(server.URL)

	// Test successful health check
	err := adapter.HealthCheck()
	if err != nil {
		t.Errorf("Expected successful health check, got error: %v", err)
	}
}

func TestDGraphGraphQLAdapter_NewDGraphGraphQLAdapter(t *testing.T) {
	endpoint := "http://localhost:8080/graphql"
	adapter := NewDGraphGraphQLAdapter(endpoint)

	if adapter == nil {
		t.Error("Expected adapter to be created, got nil")
	}

	if adapter.endpoint != endpoint {
		t.Errorf("Expected endpoint '%s', got '%s'", endpoint, adapter.endpoint)
	}

	if adapter.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestDGraphGraphQLAdapter_GetEndpoint(t *testing.T) {
	endpoint := "http://localhost:8080/graphql"
	adapter := NewDGraphGraphQLAdapter(endpoint)

	result := adapter.GetEndpoint()
	if result != endpoint {
		t.Errorf("Expected endpoint '%s', got '%s'", endpoint, result)
	}
}

func TestDGraphGraphQLAdapter_SetTimeout(t *testing.T) {
	adapter := NewDGraphGraphQLAdapter("http://localhost:8080/graphql")

	// Test setting timeout
	timeout := 60 * time.Second
	adapter.SetTimeout(timeout)

	// Verify timeout was set (we can't easily test the internal timeout, but we can verify no panic)
	if adapter.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestDGraphGraphQLAdapter_RetryLogic(t *testing.T) {
	// Create a test server that fails first, then succeeds
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Succeed on third attempt
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"queryGame": []map[string]interface{}{
					{"id": "game1", "activePlayer": "player1"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	adapter := NewDGraphGraphQLAdapter(server.URL)

	// Test that retry logic works
	result, err := adapter.Query("query GetGames { queryGame { id activePlayer } }", nil)

	if err != nil {
		t.Errorf("Expected successful query after retries, got error: %v", err)
	}

	if result == nil {
		t.Error("Expected result after successful retry")
	}

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}
}
