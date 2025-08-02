package app

import (
	"fmt"
	"time"

	"github.com/18xxnotifier/internal/adapters"
	"github.com/18xxnotifier/internal/domain/entities"
)

// GameService handles game-related business logic
type GameService struct {
	gameRepo        entities.GameRepository
	gameDataAdapter adapters.GameDataAdapter
}

// NewGameService creates a new GameService instance
func NewGameService(gameRepo entities.GameRepository, gameDataAdapter adapters.GameDataAdapter) *GameService {
	return &GameService{
		gameRepo:        gameRepo,
		gameDataAdapter: gameDataAdapter,
	}
}

// TrackGame starts tracking a new game
func (s *GameService) TrackGame(gameID string) error {
	// Fetch initial game data from external API
	game, err := s.gameDataAdapter.FetchGameData(gameID)
	if err != nil {
		return fmt.Errorf("failed to fetch game data: %w", err)
	}

	if game == nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	// Save game to repository
	err = s.gameRepo.SaveGame(game)
	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

// GetActiveGames returns all currently active games
func (s *GameService) GetActiveGames() ([]*entities.Game, error) {
	return s.gameRepo.GetActiveGames()
}

// GetGame retrieves a specific game by ID
func (s *GameService) GetGame(gameID string) (*entities.Game, error) {
	return s.gameRepo.GetGame(gameID)
}

// GetAllGames returns all games
func (s *GameService) GetAllGames() ([]*entities.Game, error) {
	return s.gameRepo.GetAllGames()
}

// UpdateGameData updates game data and detects changes
func (s *GameService) UpdateGameData(gameID string) ([]*entities.GameChange, error) {
	// Get current game data from repository
	currentGame, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current game: %w", err)
	}

	if currentGame == nil {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	// Fetch latest game data from external API
	newGame, err := s.gameDataAdapter.FetchGameData(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new game data: %w", err)
	}

	if newGame == nil {
		return nil, fmt.Errorf("game not found in external API: %s", gameID)
	}

	// Detect changes between old and new game data
	changes := s.DetectChanges(currentGame, newGame)

	// Update game in repository
	err = s.gameRepo.UpdateGame(newGame)
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return changes, nil
}

// DetectChanges compares two game states and returns detected changes
func (s *GameService) DetectChanges(oldGame, newGame *entities.Game) []*entities.GameChange {
	var changes []*entities.GameChange

	// Detect active player changes
	if oldGame.ActivePlayer != newGame.ActivePlayer {
		changes = append(changes, &entities.GameChange{
			GameID:     newGame.ID,
			ChangeType: "player_change",
			OldValue:   oldGame.ActivePlayer,
			NewValue:   newGame.ActivePlayer,
			Timestamp:  time.Now(),
		})
	}

	// Detect game end
	if !oldGame.Finished && newGame.Finished {
		changes = append(changes, &entities.GameChange{
			GameID:     newGame.ID,
			ChangeType: "game_end",
			OldValue:   "active",
			NewValue:   "finished",
			Timestamp:  time.Now(),
		})
	}

	// Detect turn start (when game becomes active)
	if oldGame.Finished && !newGame.Finished && newGame.ActivePlayer != "" {
		changes = append(changes, &entities.GameChange{
			GameID:     newGame.ID,
			ChangeType: "turn_start",
			OldValue:   "finished",
			NewValue:   "active",
			Timestamp:  time.Now(),
		})
	}

	return changes
}

// StopTrackingGame removes a game from tracking
func (s *GameService) StopTrackingGame(gameID string) error {
	// Get the game to verify it exists
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game == nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	// Delete the game from repository
	err = s.gameRepo.DeleteGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}
