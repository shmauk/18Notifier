package entities

import (
	"testing"
	"time"
)

func TestGame_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		game     *Game
		expected bool
	}{
		{
			name: "active game",
			game: &Game{
				ID:           "123",
				ActivePlayer: "player1",
				Finished:     false,
			},
			expected: true,
		},
		{
			name: "finished game",
			game: &Game{
				ID:           "123",
				ActivePlayer: "",
				Finished:     true,
			},
			expected: false,
		},
		{
			name: "no active player",
			game: &Game{
				ID:           "123",
				ActivePlayer: "",
				Finished:     false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.IsActive()
			if result != tt.expected {
				t.Errorf("Game.IsActive() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGame_GetLastUpdated(t *testing.T) {
	now := time.Now()
	game := &Game{
		ID:           "123",
		LastUpdated:  now,
		ActivePlayer: "player1",
		Finished:     false,
	}

	result := game.GetLastUpdated()
	if !result.Equal(now) {
		t.Errorf("Game.GetLastUpdated() = %v, want %v", result, now)
	}
}

func TestGame_GetPlayers(t *testing.T) {
	game := &Game{
		ID:           "123",
		Players:      []string{"player1", "player2", "player3"},
		ActivePlayer: "player1",
		Finished:     false,
	}

	result := game.GetPlayers()
	expected := []string{"player1", "player2", "player3"}

	if len(result) != len(expected) {
		t.Errorf("Game.GetPlayers() length = %d, want %d", len(result), len(expected))
	}

	for i, player := range result {
		if player != expected[i] {
			t.Errorf("Game.GetPlayers()[%d] = %s, want %s", i, player, expected[i])
		}
	}
}

func TestGame_GetActivePlayer(t *testing.T) {
	game := &Game{
		ID:           "123",
		ActivePlayer: "player1",
		Finished:     false,
	}

	result := game.GetActivePlayer()
	expected := "player1"

	if result != expected {
		t.Errorf("Game.GetActivePlayer() = %s, want %s", result, expected)
	}
}

func TestGame_IsFinished(t *testing.T) {
	tests := []struct {
		name     string
		game     *Game
		expected bool
	}{
		{
			name: "finished game",
			game: &Game{
				ID:       "123",
				Finished: true,
			},
			expected: true,
		},
		{
			name: "active game",
			game: &Game{
				ID:       "123",
				Finished: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.IsFinished()
			if result != tt.expected {
				t.Errorf("Game.IsFinished() = %v, want %v", result, tt.expected)
			}
		})
	}
}
