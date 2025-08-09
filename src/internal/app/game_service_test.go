package app

import (
	"strings"
	"testing"
	"time"

	"github.com/18xxnotifier/internal/domain/entities"
)

// MockGameRepository implements entities.GameRepository for testing
type MockGameRepository struct {
	games map[string]*entities.Game
}

func NewMockGameRepository() *MockGameRepository {
	return &MockGameRepository{
		games: make(map[string]*entities.Game),
	}
}

func (m *MockGameRepository) GetGame(id string) (*entities.Game, error) {
	if game, exists := m.games[id]; exists {
		return game, nil
	}
	return nil, nil
}

func (m *MockGameRepository) SaveGame(game *entities.Game, channelID string, guildID string) error {
	m.games[game.ID] = game
	return nil
}

func (m *MockGameRepository) UpdateGame(game *entities.Game) error {
	m.games[game.ID] = game
	return nil
}

func (m *MockGameRepository) DeleteGame(id string) error {
	delete(m.games, id)
	return nil
}

func (m *MockGameRepository) GetAllGames() ([]*entities.Game, error) {
	var games []*entities.Game
	for _, game := range m.games {
		games = append(games, game)
	}
	return games, nil
}

func (m *MockGameRepository) GetActiveGames() ([]*entities.Game, error) {
	var games []*entities.Game
	for _, game := range m.games {
		if game.IsActive() {
			games = append(games, game)
		}
	}
	return games, nil
}

// MockGameDataAdapter implements adapters.GameDataAdapter for testing
type MockGameDataAdapter struct {
	games map[string]*entities.Game
}

func NewMockGameDataAdapter() *MockGameDataAdapter {
	return &MockGameDataAdapter{
		games: make(map[string]*entities.Game),
	}
}

func (m *MockGameDataAdapter) GetGameData(gameID string) (*entities.Game, error) {
	if game, exists := m.games[gameID]; exists {
		return game, nil
	}
	return nil, nil
}

func (m *MockGameDataAdapter) SetGameData(gameID string, game *entities.Game) {
	m.games[gameID] = game
}

func (m *MockGameDataAdapter) FetchGameData(gameID string) (*entities.Game, error) {
	if game, exists := m.games[gameID]; exists {
		return game, nil
	}
	return nil, nil
}

func TestGameService_TrackGame(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Test successful tracking
	gameID := "test-game-123"
	game := &entities.Game{
		ID:            gameID,
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{"player1"},
		Finished:      false,
		LastUpdated:   time.Now(),
	}
	adapter.SetGameData(gameID, game)

	err := service.TrackGame(gameID, "test-channel", "test-guild")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	savedGame, err := repo.GetGame(gameID)
	if err != nil {
		t.Errorf("Expected no error getting saved game, got %v", err)
	}
	if savedGame == nil {
		t.Error("Expected saved game to exist")
	}
	if savedGame.ID != gameID {
		t.Errorf("Expected game ID %s, got %s", gameID, savedGame.ID)
	}
}

func TestGameService_TrackGame_GameNotFound(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Test tracking non-existent game
	err := service.TrackGame("non-existent-game", "test-channel", "test-guild")
	if err == nil {
		t.Error("Expected error for non-existent game")
	}
	if !strings.Contains(err.Error(), "game not found") {
		t.Errorf("Expected 'game not found' error, got %v", err)
	}
}

func TestGameService_GetActiveGames(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Add some test games
	activeGame := &entities.Game{
		ID:            "active-game",
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{"player1"},
		Finished:      false,
		LastUpdated:   time.Now(),
	}
	finishedGame := &entities.Game{
		ID:            "finished-game",
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{},
		Finished:      true,
		LastUpdated:   time.Now(),
	}

	repo.SaveGame(activeGame, "test-channel", "test-guild")
	repo.SaveGame(finishedGame, "test-channel", "test-guild")

	games, err := service.GetActiveGames()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(games) != 1 {
		t.Errorf("Expected 1 active game, got %d", len(games))
	}

	if games[0].ID != "active-game" {
		t.Errorf("Expected active game ID 'active-game', got %s", games[0].ID)
	}
}

func TestGameService_UpdateGameData(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	gameID := "test-game-123"
	oldGame := &entities.Game{
		ID:            gameID,
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{"player1"},
		Finished:      false,
		LastUpdated:   time.Now(),
	}
	newGame := &entities.Game{
		ID:            gameID,
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{"player2"}, // Changed active player
		Finished:      false,
		LastUpdated:   time.Now(),
	}

	repo.SaveGame(oldGame, "test-channel", "test-guild")
	adapter.SetGameData(gameID, newGame)

	changes, err := service.UpdateGameData(gameID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	if change.ChangeType != "player_change" {
		t.Errorf("Expected change type 'player_change', got %s", change.ChangeType)
	}
	if change.OldValue != "player1" {
		t.Errorf("Expected old value 'player1', got %s", change.OldValue)
	}
	if change.NewValue != "player2" {
		t.Errorf("Expected new value 'player2', got %s", change.NewValue)
	}
}

func TestGameService_StopTrackingGame(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	gameID := "test-game-123"
	game := &entities.Game{
		ID:            gameID,
		Players:       []string{"player1", "player2"},
		ActivePlayers: []string{"player1"},
		Finished:      false,
		LastUpdated:   time.Now(),
	}

	repo.SaveGame(game, "test-channel", "test-guild")

	err := service.StopTrackingGame(gameID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	savedGame, err := repo.GetGame(gameID)
	if err != nil {
		t.Errorf("Expected no error getting game, got %v", err)
	}
	if savedGame != nil {
		t.Error("Expected game to be deleted")
	}
}

func TestGameService_StopTrackingGame_GameNotFound(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	err := service.StopTrackingGame("non-existent-game")
	if err == nil {
		t.Error("Expected error for non-existent game")
	}
	if !strings.Contains(err.Error(), "game not found") {
		t.Errorf("Expected 'game not found' error, got %v", err)
	}
}
