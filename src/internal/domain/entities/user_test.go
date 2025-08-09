package entities

import (
	"testing"
)

func TestUser_HasAccount(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		account  string
		expected bool
	}{
		{
			name: "user has account",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{"player1", "player2"},
				SubscribedGames:    []string{"game1"},
			},
			account:  "player1",
			expected: true,
		},
		{
			name: "user does not have account",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{"player1", "player2"},
				SubscribedGames:    []string{"game1"},
			},
			account:  "player3",
			expected: false,
		},
		{
			name: "user has no accounts",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{},
				SubscribedGames:    []string{},
			},
			account:  "player1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasAccount(tt.account)
			if result != tt.expected {
				t.Errorf("User.HasAccount(%s) = %v, want %v", tt.account, result, tt.expected)
			}
		})
	}
}

func TestUser_IsSubscribedToGame(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		gameID   string
		expected bool
	}{
		{
			name: "user is subscribed to game",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{"player1"},
				SubscribedGames:    []string{"game1", "game2"},
			},
			gameID:   "game1",
			expected: true,
		},
		{
			name: "user is not subscribed to game",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{"player1"},
				SubscribedGames:    []string{"game1", "game2"},
			},
			gameID:   "game3",
			expected: false,
		},
		{
			name: "user has no subscriptions",
			user: &User{
				DiscordID:          "123456789",
				EighteenxxAccounts: []string{"player1"},
				SubscribedGames:    []string{},
			},
			gameID:   "game1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsSubscribedToGame(tt.gameID)
			if result != tt.expected {
				t.Errorf("User.IsSubscribedToGame(%s) = %v, want %v", tt.gameID, result, tt.expected)
			}
		})
	}
}

func TestUser_AddAccount(t *testing.T) {
	user := &User{
		DiscordID:          "123456789",
		EighteenxxAccounts: []string{"player1"},
		SubscribedGames:    []string{"game1"},
	}

	// Test adding new account
	user.AddAccount("player2")
	expected := []string{"player1", "player2"}

	if len(user.EighteenxxAccounts) != len(expected) {
		t.Errorf("User.AddAccount() length = %d, want %d", len(user.EighteenxxAccounts), len(expected))
	}

	for i, account := range user.EighteenxxAccounts {
		if account != expected[i] {
			t.Errorf("User.AddAccount()[%d] = %s, want %s", i, account, expected[i])
		}
	}

	// Test adding duplicate account (should not add)
	user.AddAccount("player1")
	if len(user.EighteenxxAccounts) != len(expected) {
		t.Errorf("User.AddAccount() duplicate length = %d, want %d", len(user.EighteenxxAccounts), len(expected))
	}
}

func TestUser_RemoveAccount(t *testing.T) {
	user := &User{
		DiscordID:          "123456789",
		EighteenxxAccounts: []string{"player1", "player2", "player3"},
		SubscribedGames:    []string{"game1"},
	}

	// Test removing existing account
	user.RemoveAccount("player2")
	expected := []string{"player1", "player3"}

	if len(user.EighteenxxAccounts) != len(expected) {
		t.Errorf("User.RemoveAccount() length = %d, want %d", len(user.EighteenxxAccounts), len(expected))
	}

	for i, account := range user.EighteenxxAccounts {
		if account != expected[i] {
			t.Errorf("User.RemoveAccount()[%d] = %s, want %s", i, account, expected[i])
		}
	}

	// Test removing non-existent account
	user.RemoveAccount("player4")
	if len(user.EighteenxxAccounts) != len(expected) {
		t.Errorf("User.RemoveAccount() non-existent length = %d, want %d", len(user.EighteenxxAccounts), len(expected))
	}
}

func TestUser_SubscribeToGame(t *testing.T) {
	user := &User{
		DiscordID:          "123456789",
		EighteenxxAccounts: []string{"player1"},
		SubscribedGames:    []string{"game1"},
	}

	// Test subscribing to new game
	user.SubscribeToGame("game2")
	expected := []string{"game1", "game2"}

	if len(user.SubscribedGames) != len(expected) {
		t.Errorf("User.SubscribeToGame() length = %d, want %d", len(user.SubscribedGames), len(expected))
	}

	for i, game := range user.SubscribedGames {
		if game != expected[i] {
			t.Errorf("User.SubscribeToGame()[%d] = %s, want %s", i, game, expected[i])
		}
	}

	// Test subscribing to existing game (should not add)
	user.SubscribeToGame("game1")
	if len(user.SubscribedGames) != len(expected) {
		t.Errorf("User.SubscribeToGame() duplicate length = %d, want %d", len(user.SubscribedGames), len(expected))
	}
}

func TestUser_UnsubscribeFromGame(t *testing.T) {
	user := &User{
		DiscordID:          "123456789",
		EighteenxxAccounts: []string{"player1"},
		SubscribedGames:    []string{"game1", "game2", "game3"},
	}

	// Test unsubscribing from existing game
	user.UnsubscribeFromGame("game2")
	expected := []string{"game1", "game3"}

	if len(user.SubscribedGames) != len(expected) {
		t.Errorf("User.UnsubscribeFromGame() length = %d, want %d", len(user.SubscribedGames), len(expected))
	}

	for i, game := range user.SubscribedGames {
		if game != expected[i] {
			t.Errorf("User.UnsubscribeFromGame()[%d] = %s, want %s", i, game, expected[i])
		}
	}

	// Test unsubscribing from non-existent game
	user.UnsubscribeFromGame("game4")
	if len(user.SubscribedGames) != len(expected) {
		t.Errorf("User.UnsubscribeFromGame() non-existent length = %d, want %d", len(user.SubscribedGames), len(expected))
	}
}
