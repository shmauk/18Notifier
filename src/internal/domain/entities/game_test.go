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
				ID:            "123",
				ActivePlayers: []string{"player1"},
				Finished:      false,
			},
			expected: true,
		},
		{
			name: "finished game",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{},
				Finished:      true,
			},
			expected: false,
		},
		{
			name: "no active player",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{},
				Finished:      false,
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
		ID:            "123",
		LastUpdated:   now,
		ActivePlayers: []string{"player1"},
		Finished:      false,
	}

	result := game.GetLastUpdated()
	if !result.Equal(now) {
		t.Errorf("Game.GetLastUpdated() = %v, want %v", result, now)
	}
}

func TestGame_GetPlayers(t *testing.T) {
	game := &Game{
		ID:            "123",
		Players:       []string{"player1", "player2", "player3"},
		ActivePlayers: []string{"player1"},
		Finished:      false,
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
	tests := []struct {
		name     string
		game     *Game
		expected string
	}{
		{
			name: "single active player",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{"player1"},
				Finished:      false,
			},
			expected: "player1",
		},
		{
			name: "multiple active players",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{"player1", "player2"},
				Finished:      false,
			},
			expected: "player1", // Should return first player
		},
		{
			name: "no active players",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{},
				Finished:      false,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.GetActivePlayer()
			if result != tt.expected {
				t.Errorf("Game.GetActivePlayer() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGame_GetActivePlayers(t *testing.T) {
	tests := []struct {
		name     string
		game     *Game
		expected []string
	}{
		{
			name: "single active player",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{"player1"},
				Finished:      false,
			},
			expected: []string{"player1"},
		},
		{
			name: "multiple active players",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{"player1", "player2", "player3"},
				Finished:      false,
			},
			expected: []string{"player1", "player2", "player3"},
		},
		{
			name: "no active players",
			game: &Game{
				ID:            "123",
				ActivePlayers: []string{},
				Finished:      false,
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.game.GetActivePlayers()

			if len(result) != len(tt.expected) {
				t.Errorf("Game.GetActivePlayers() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i, player := range result {
				if player != tt.expected[i] {
					t.Errorf("Game.GetActivePlayers()[%d] = %s, want %s", i, player, tt.expected[i])
				}
			}
		})
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
