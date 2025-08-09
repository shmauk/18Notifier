package entities

import (
	"testing"
)

func TestChannel_HasGame(t *testing.T) {
	tests := []struct {
		name     string
		channel  *Channel
		gameID   string
		expected bool
	}{
		{
			name: "channel has game",
			channel: &Channel{
				ID:      "channel1",
				GuildID: "guild1",
				Name:    "general",
				Games:   []string{"game1", "game2"},
			},
			gameID:   "game1",
			expected: true,
		},
		{
			name: "channel does not have game",
			channel: &Channel{
				ID:      "channel1",
				GuildID: "guild1",
				Name:    "general",
				Games:   []string{"game1", "game2"},
			},
			gameID:   "game3",
			expected: false,
		},
		{
			name: "channel has no games",
			channel: &Channel{
				ID:      "channel1",
				GuildID: "guild1",
				Name:    "general",
				Games:   []string{},
			},
			gameID:   "game1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.channel.HasGame(tt.gameID)
			if result != tt.expected {
				t.Errorf("Channel.HasGame(%s) = %v, want %v", tt.gameID, result, tt.expected)
			}
		})
	}
}

func TestChannel_AddGame(t *testing.T) {
	channel := &Channel{
		ID:      "channel1",
		GuildID: "guild1",
		Name:    "general",
		Games:   []string{"game1"},
	}

	// Test adding new game
	channel.AddGame("game2")
	expected := []string{"game1", "game2"}

	if len(channel.Games) != len(expected) {
		t.Errorf("Channel.AddGame() length = %d, want %d", len(channel.Games), len(expected))
	}

	for i, game := range channel.Games {
		if game != expected[i] {
			t.Errorf("Channel.AddGame()[%d] = %s, want %s", i, game, expected[i])
		}
	}

	// Test adding duplicate game (should not add)
	channel.AddGame("game1")
	if len(channel.Games) != len(expected) {
		t.Errorf("Channel.AddGame() duplicate length = %d, want %d", len(channel.Games), len(expected))
	}
}

func TestChannel_RemoveGame(t *testing.T) {
	channel := &Channel{
		ID:      "channel1",
		GuildID: "guild1",
		Name:    "general",
		Games:   []string{"game1", "game2", "game3"},
	}

	// Test removing existing game
	channel.RemoveGame("game2")
	expected := []string{"game1", "game3"}

	if len(channel.Games) != len(expected) {
		t.Errorf("Channel.RemoveGame() length = %d, want %d", len(channel.Games), len(expected))
	}

	for i, game := range channel.Games {
		if game != expected[i] {
			t.Errorf("Channel.RemoveGame()[%d] = %s, want %s", i, game, expected[i])
		}
	}

	// Test removing non-existent game
	channel.RemoveGame("game4")
	if len(channel.Games) != len(expected) {
		t.Errorf("Channel.RemoveGame() non-existent length = %d, want %d", len(channel.Games), len(expected))
	}
}

func TestChannel_GetGames(t *testing.T) {
	channel := &Channel{
		ID:      "channel1",
		GuildID: "guild1",
		Name:    "general",
		Games:   []string{"game1", "game2", "game3"},
	}

	result := channel.GetGames()
	expected := []string{"game1", "game2", "game3"}

	if len(result) != len(expected) {
		t.Errorf("Channel.GetGames() length = %d, want %d", len(result), len(expected))
	}

	for i, game := range result {
		if game != expected[i] {
			t.Errorf("Channel.GetGames()[%d] = %s, want %s", i, game, expected[i])
		}
	}
}

func TestChannel_GetName(t *testing.T) {
	channel := &Channel{
		ID:      "channel1",
		GuildID: "guild1",
		Name:    "general",
		Games:   []string{"game1"},
	}

	result := channel.GetName()
	expected := "general"

	if result != expected {
		t.Errorf("Channel.GetName() = %s, want %s", result, expected)
	}
}

func TestChannel_GetGuildID(t *testing.T) {
	channel := &Channel{
		ID:      "channel1",
		GuildID: "guild1",
		Name:    "general",
		Games:   []string{"game1"},
	}

	result := channel.GetGuildID()
	expected := "guild1"

	if result != expected {
		t.Errorf("Channel.GetGuildID() = %s, want %s", result, expected)
	}
}
