package app

import (
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

func (m *MockGameRepository) SaveGame(game *entities.Game) error {
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

	// Set up test game data in the adapter
	gameID := "test-game-123"
	testGame := &entities.Game{
		ID:           gameID,
		ActivePlayer: "player1",
		Finished:     false,
		LastUpdated:  time.Now(),
	}
	adapter.SetGameData(gameID, testGame)

	// Test tracking a new game
	err := service.TrackGame(gameID)
	if err != nil {
		t.Errorf("GameService.TrackGame() error = %v", err)
	}

	// Verify game was saved
	game, err := repo.GetGame(gameID)
	if err != nil {
		t.Errorf("Failed to get tracked game: %v", err)
	}
	if game == nil {
		t.Error("Tracked game not found in repository")
	}
}

func TestGameService_GetActiveGames(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Add some test games
	activeGame := &entities.Game{
		ID:           "active-game",
		ActivePlayer: "player1",
		Finished:     false,
		LastUpdated:  time.Now(),
	}
	finishedGame := &entities.Game{
		ID:           "finished-game",
		ActivePlayer: "",
		Finished:     true,
		LastUpdated:  time.Now(),
	}

	repo.SaveGame(activeGame)
	repo.SaveGame(finishedGame)

	// Test getting active games
	games, err := service.GetActiveGames()
	if err != nil {
		t.Errorf("GameService.GetActiveGames() error = %v", err)
	}

	if len(games) != 1 {
		t.Errorf("Expected 1 active game, got %d", len(games))
	}

	if games[0].ID != "active-game" {
		t.Errorf("Expected active game ID 'active-game', got '%s'", games[0].ID)
	}
}

func TestGameService_UpdateGameData(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Set up test data
	gameID := "test-game"
	oldGame := &entities.Game{
		ID:           gameID,
		ActivePlayer: "player1",
		Finished:     false,
		LastUpdated:  time.Now().Add(-time.Hour),
	}
	newGame := &entities.Game{
		ID:           gameID,
		ActivePlayer: "player2",
		Finished:     false,
		LastUpdated:  time.Now(),
	}

	// Save old game and set up adapter
	repo.SaveGame(oldGame)
	adapter.SetGameData(gameID, newGame)

	// Test updating game data
	changes, err := service.UpdateGameData(gameID)
	if err != nil {
		t.Errorf("GameService.UpdateGameData() error = %v", err)
	}

	// Verify changes were detected
	if len(changes) == 0 {
		t.Error("Expected changes to be detected, but none were found")
	}

	// Verify game was updated in repository
	updatedGame, err := repo.GetGame(gameID)
	if err != nil {
		t.Errorf("Failed to get updated game: %v", err)
	}
	if updatedGame.ActivePlayer != "player2" {
		t.Errorf("Expected active player 'player2', got '%s'", updatedGame.ActivePlayer)
	}
}

func TestGameService_GetGame(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Add test game
	game := &entities.Game{
		ID:           "test-game",
		ActivePlayer: "player1",
		Finished:     false,
		LastUpdated:  time.Now(),
	}
	repo.SaveGame(game)

	// Test getting game
	retrievedGame, err := service.GetGame("test-game")
	if err != nil {
		t.Errorf("GameService.GetGame() error = %v", err)
	}

	if retrievedGame == nil {
		t.Error("Expected game to be retrieved, got nil")
	}

	if retrievedGame.ID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", retrievedGame.ID)
	}
}

func TestGameService_GetAllGames(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Add test games
	game1 := &entities.Game{ID: "game1", ActivePlayer: "player1", Finished: false}
	game2 := &entities.Game{ID: "game2", ActivePlayer: "player2", Finished: false}
	repo.SaveGame(game1)
	repo.SaveGame(game2)

	// Test getting all games
	games, err := service.GetAllGames()
	if err != nil {
		t.Errorf("GameService.GetAllGames() error = %v", err)
	}

	if len(games) != 2 {
		t.Errorf("Expected 2 games, got %d", len(games))
	}
}

func TestGameService_DetectChanges(t *testing.T) {
	repo := NewMockGameRepository()
	adapter := NewMockGameDataAdapter()
	service := NewGameService(repo, adapter)

	// Set up test data
	gameID := "test-game"
	oldGame := &entities.Game{
		ID:           gameID,
		ActivePlayer: "player1",
		Finished:     false,
		LastUpdated:  time.Now().Add(-time.Hour),
	}
	newGame := &entities.Game{
		ID:           gameID,
		ActivePlayer: "player2",
		Finished:     false,
		LastUpdated:  time.Now(),
	}

	// Test change detection
	changes := service.DetectChanges(oldGame, newGame)

	// Verify changes were detected
	if len(changes) == 0 {
		t.Error("Expected changes to be detected, but none were found")
	}

	// Verify specific change
	foundPlayerChange := false
	for _, change := range changes {
		if change.ChangeType == "player_change" {
			foundPlayerChange = true
			if change.OldValue != "player1" || change.NewValue != "player2" {
				t.Errorf("Expected player change from 'player1' to 'player2', got '%s' to '%s'",
					change.OldValue, change.NewValue)
			}
		}
	}

	if !foundPlayerChange {
		t.Error("Expected player change to be detected")
	}
}
