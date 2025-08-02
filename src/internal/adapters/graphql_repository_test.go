package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGraphQLRepository_Query(t *testing.T) {
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
		case "query GetGame($id: String!) { getGame(id: $id) { id activePlayer finished } }":
			response = map[string]interface{}{
				"data": map[string]interface{}{
					"getGame": map[string]interface{}{
						"id":           "game1",
						"activePlayer": "player1",
						"finished":     false,
					},
				},
			}
		case "query QueryGame { queryGame { id activePlayer finished } }":
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

	// Create repository with test server URL
	repo := &GraphQLGameRepository{
		graphqlAdapter: NewDGraphGraphQLAdapter(server.URL),
	}

	tests := []struct {
		name     string
		query    string
		variables map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "get game by ID",
			query:    "query GetGame($id: String!) { getGame(id: $id) { id activePlayer finished } }",
			variables: map[string]interface{}{"id": "game1"},
			wantErr:  false,
		},
		{
			name:     "query all games",
			query:    "query QueryGame { queryGame { id activePlayer finished } }",
			variables: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.graphqlAdapter.Query(tt.query, tt.variables)

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

func TestGraphQLRepository_Mutate(t *testing.T) {
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

	// Create repository with test server URL
	repo := &GraphQLGameRepository{
		graphqlAdapter: NewDGraphGraphQLAdapter(server.URL),
	}

	tests := []struct {
		name     string
		mutation string
		variables map[string]interface{}
		wantErr  bool
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
			result, err := repo.graphqlAdapter.Mutate(tt.mutation, tt.variables)

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



func TestGraphQLRepository_NewGraphQLGameRepository(t *testing.T) {
	graphqlAdapter := NewDGraphGraphQLAdapter("http://localhost:8080/graphql")
	repo := NewGraphQLGameRepository(graphqlAdapter)

	if repo == nil {
		t.Error("Expected repository to be created, got nil")
	}

	if repo.graphqlAdapter == nil {
		t.Error("Expected GraphQL adapter to be initialized")
	}
} 